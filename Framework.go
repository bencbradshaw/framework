package framework

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/evanw/esbuild/pkg/api"

	"github.com/bencbradshaw/framework/env"
	"github.com/bencbradshaw/framework/esbuild"
	"github.com/bencbradshaw/framework/events"
	"github.com/bencbradshaw/framework/twig"
)

type RouterSetupFunc struct {
	BasePath string
	Handler  func(mux *http.ServeMux, db interface{}, devMode bool) http.Handler
}

type InitParams struct {
	Mux                        *http.ServeMux
	IsDevMode                  bool
	EsbuildOpts                api.BuildOptions
	AutoRegisterTemplateRoutes bool
}

func Render(w http.ResponseWriter, name string, data map[string]interface{}) {
	result, err := twig.Render(name, data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(result))
}

func Run(params InitParams) *http.ServeMux {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("Error determining current file directory")
	}
	currentDir := filepath.Dir(filename)

	_, err := env.LoadEnvFile(filepath.Join(currentDir, ".env"))
	if err != nil {
		log.Printf("No .env file loaded: %v", err)
	}

	devMode := params.IsDevMode
	if !params.IsDevMode {
		flag.BoolVar(&devMode, "dev", false, "Run in development mode")
		flag.Parse()
	}

	fmt.Println("Running in dev mode:", devMode)

	if devMode {
		esbuild.InitDevMode(params.EsbuildOpts)
		print("Dev mode initialized \n")
	}

	var mux *http.ServeMux
	if params.Mux == nil {
		mux = http.NewServeMux()
	} else {
		mux = params.Mux
	}
	autoRegisterTemplateRoutes := params.AutoRegisterTemplateRoutes
	if autoRegisterTemplateRoutes {
		templateDir := "templates"
		err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(info.Name()) == ".twig" && strings.Contains(info.Name(), "route") {
				tmplName := info.Name()
				baseName := strings.Split(tmplName, ".")[0]
				routePath := "/" + baseName
				if strings.Contains(tmplName, "subroute") {
					routePath += "/"
				}
				mux.HandleFunc(routePath, func(w http.ResponseWriter, r *http.Request) {
					Render(
						w,
						tmplName,
						map[string]interface{}{"title": tmplName},
					)
				})
				fmt.Printf("Registered route for template: %s -> %s\n", tmplName, routePath)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking through templates directory: %v", err)
		}
		mux.HandleFunc("/events", events.EventStream)

		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			Render(
				w,
				"index.twig",
				map[string]interface{}{"title": "home"},
			)
		})
	}

	return mux
}

func Build(params InitParams) api.BuildResult {
	return esbuild.Build(params.EsbuildOpts)
}
