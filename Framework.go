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
}

func RenderWithHtmlResponse(w http.ResponseWriter, templateName string, data map[string]interface{}) {
	fmt.Println("Rendering template: ", templateName)

	result, err := templating.HtmlRender(templateName, data)
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

	devMode := params.IsDevMode // Initialize from params
	if !params.IsDevMode {      // If mode isn't forced by params, consider flags
		if flag.Lookup("dev") == nil {
			// Note: `devMode` here is the local variable. The flag will bind to its address
			// for the current call to Run. This is generally fine for controlling `devMode`
			// within this specific Run invocation based on command-line flags,
			// but doesn't create a persistent global flag state that other packages might inspect
			// via the flag package directly (unless they lookup after this Run call parsed).
			flag.BoolVar(&devMode, "dev", false, "Run in development mode")
		}

		if !flag.Parsed() {
			flag.Parse()
		}

		// After parsing (or if already parsed), ensure `devMode` reflects the actual state of the "dev" flag.
		// This is crucial if `flag.Parse()` was called in a previous test or Run invocation,
		// or if tests are run with `-args -dev`.
		if f := flag.Lookup("dev"); f != nil { // Check if the "dev" flag is defined
			// Access the flag's value. For a BoolVar, it's true if "true", false otherwise.
			if f.Value.String() == "true" {
				devMode = true
			} else {
				devMode = false // Catches "false" or any other non-"true" string for the bool flag
			}
		} else {
			// If the "dev" flag was never defined (e.g. Lookup was nil and we didn't define it),
			// and params.IsDevMode is false, then devMode should be false.
			// This is already handled as devMode was initialized to params.IsDevMode (false in this block).
			// So, devMode = false; explicitly.
			devMode = false
		}
	}

	fmt.Println("Running in dev mode:", devMode)

	if devMode {
		esbuild.InitDevMode(params.EsbuildOpts)
		print("Dev mode initialized \n")
	}

	mux := http.NewServeMux()

	if params.AutoRegisterTemplateRoutes {
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

				mux.HandleFunc(routePath, func(w http.ResponseWriter, r *http.Request) {
					fmt.Printf("handling request for route: %s\n", routePath)
					RenderWithHtmlResponse(
						w,
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

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			RenderWithHtmlResponse(
				w,
				"index.html",
				map[string]any{"title": "Home"},
			)
		})
	}

	mux.HandleFunc("/events", events.EventStream)

	// Determine project root to make static file path absolute
	// Assuming Framework.go is at project root. If it can be deeper, this needs adjustment.
	// For this project, Framework.go is at the root.
	_, mainGoFilename, _, ok := runtime.Caller(0) // Get path of Framework.go
	if !ok {
		log.Fatalf("Could not get current file path for static dir setup")
	}
	projectRootDir := filepath.Dir(mainGoFilename)
	staticDir := filepath.Join(projectRootDir, "static")
	log.Printf("Serving static files from: %s", staticDir)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	return mux
}

func Build(params InitParams) api.BuildResult {
	return esbuild.Build(params.EsbuildOpts)
}
