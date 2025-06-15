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
	"github.com/bencbradshaw/framework/templating"
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
	TemplatesDir               string
}

// RenderWithHtmlResponse renders the specified HTML template with given data and writes it to the response.
// It now accepts templatesDir as its second argument.
func RenderWithHtmlResponse(w http.ResponseWriter, templatesDir string, templateName string, data map[string]interface{}) {
	fmt.Println("Rendering template: ", templateName, "using dir:", templatesDir)

	result, err := templating.HtmlRender(templatesDir, templateName, data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(result))
}

func Run(params *InitParams) *http.ServeMux {
	if params == nil {
		params = &InitParams{
			IsDevMode: true,
			EsbuildOpts: api.BuildOptions{
				EntryPoints: []string{"./frontend/src/index.ts"},
			},
			AutoRegisterTemplateRoutes: true,
		}
	}

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

	// Determine templateDir once
	templateDir := params.TemplatesDir
	if templateDir == "" {
		templateDir = "templates"
	}

	if devMode {
		// Ensure EsbuildOpts has Plugins initialized
		if params.EsbuildOpts.Plugins == nil {
			params.EsbuildOpts.Plugins = []api.Plugin{}
		}
		params.EsbuildOpts.Plugins = append(params.EsbuildOpts.Plugins, esbuild.NewHtmlPlugin(templateDir))
		esbuild.InitDevMode(params.EsbuildOpts)
		print("Dev mode initialized \n")
	}

	mux := http.NewServeMux()

	if params.AutoRegisterTemplateRoutes {
		// Use the templateDir defined above. It's captured by the handler functions below.
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

				// currentTmplDir captured here for the handler
				currentTmplDir := templateDir
				mux.HandleFunc(routePath, func(w http.ResponseWriter, r *http.Request) {
					fmt.Printf("handling request for route: %s\n", routePath)
					RenderWithHtmlResponse(
						w,
						currentTmplDir, // Pass the captured templatesDir
						tmplName,
						map[string]any{"title": baseName},
					)
				})
				fmt.Printf("Registered route for template: %s -> %s\n", tmplName, routePath)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking through templates directory: %v", err)
		}

		// currentTmplDirForRoot captured here for the root handler
		currentTmplDirForRoot := templateDir
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			RenderWithHtmlResponse(
				w,
				currentTmplDirForRoot, // Pass the captured templatesDir
				"index.html",
				map[string]any{"title": "Home"},
			)
		})
	}

	mux.HandleFunc("/events", events.EventStream)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return mux
}

func Build(params InitParams) api.BuildResult {
	templateDir := params.TemplatesDir
	if templateDir == "" {
		templateDir = "templates"
	}
	// Ensure EsbuildOpts has Plugins initialized
	if params.EsbuildOpts.Plugins == nil {
		params.EsbuildOpts.Plugins = []api.Plugin{}
	}
	params.EsbuildOpts.Plugins = append(params.EsbuildOpts.Plugins, esbuild.NewHtmlPlugin(templateDir))
	return esbuild.Build(params.EsbuildOpts)
}
