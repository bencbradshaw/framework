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
	"github.com/bencbradshaw/framework/internal"
	"github.com/bencbradshaw/framework/templating"
)

type RouterSetupFunc struct {
	BasePath string
	Handler  func(mux *http.ServeMux, db any, devMode bool) http.Handler
}

type InitParams = internal.InitParams

func RenderWithHtmlResponse(w http.ResponseWriter, templateName string, data map[string]any) {
	fmt.Println("Rendering template: ", templateName)

	result, err := templating.HtmlRender(templateName, data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(result))
}

func Run(params InitParams) *http.ServeMux {
	finalParams := internal.MergeDefaults(params)

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("Error determining current file directory")
	}
	currentDir := filepath.Dir(filename)

	_, err := env.LoadEnvFile(filepath.Join(currentDir, ".env"))
	if err != nil {
		log.Printf("No .env file loaded: %v", err)
	}

	devMode := finalParams.IsDevMode
	if !finalParams.IsDevMode {
		flag.BoolVar(&devMode, "dev", false, "Run in development mode")
		flag.Parse()
	}

	fmt.Println("Running in dev mode:", devMode)

	if devMode {
		esbuild.InitDevMode(finalParams.EsbuildOpts)
		print("Dev mode initialized \n")
	}

	mux := http.NewServeMux()

	if finalParams.AutoRegisterTemplateRoutes {
		templateDir := "templates"
		err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") && !strings.Contains(info.Name(), ".custom.html") {
				tmplName := info.Name()
				baseName := strings.Split(tmplName, ".")[0]
				routePath := "/" + baseName
				if strings.Contains(tmplName, "subroute") {
					routePath += "/"
				}
				handlerFunc := func(w http.ResponseWriter, r *http.Request) {
					fmt.Printf("handling request for route: %s\n", routePath)
					RenderWithHtmlResponse(
						w,
						tmplName,
						map[string]any{"title": baseName},
					)
				}

				if finalParams.AuthGuard != nil && strings.Contains(tmplName, ".auth.") {
					mux.Handle(routePath, finalParams.AuthGuard(http.HandlerFunc(handlerFunc)))
				} else {
					mux.HandleFunc(routePath, handlerFunc)
				}

				fmt.Printf("Registered route for template: %s -> %s\n", tmplName, routePath)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking through templates directory: %v", err)
		}

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			RenderWithHtmlResponse(
				w,
				"index.html",
				map[string]any{"title": "Home"},
			)
		})
	}

	mux.HandleFunc("/events", events.EventStream)

	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Static file requested (after prefix strip): %s", r.URL.Path)
		serveDir := filepath.Join(currentDir, finalParams.StaticDir)
		log.Printf("Serving from directory: %s", serveDir)
		http.FileServer(http.Dir(serveDir)).ServeHTTP(w, r)
	})))
	return mux
}

func Build(params InitParams) api.BuildResult {
	return esbuild.Build(params.EsbuildOpts)
}
