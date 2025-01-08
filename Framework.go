package framework

import (
	"flag"
	"fmt"
	"framework/esbuild"
	"framework/events"
	"framework/middleware"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/flosch/pongo2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initPG() (*gorm.DB, error) {
	dsn := "host=" + os.Getenv("host") +
		" user=" + os.Getenv("user") +
		" password=" + os.Getenv("password") +
		" dbname=" + os.Getenv("dbname") +
		" port=" + os.Getenv("port") +
		" sslmode=disable application_name=novaopus"
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

type RouterSetupFunc struct {
	BasePath string
	Handler  func(mux *http.ServeMux, db interface{}, devMode bool) http.Handler
}

type InitParams struct {
	Mux                        *http.ServeMux
	IsDevMode                  bool
	EsbuildOpts                api.BuildOptions
	RouterSetupFuncs           []RouterSetupFunc
	DB                         interface{}
	AutoRegisterTemplateRoutes bool
}

func Render(w http.ResponseWriter, name string, data pongo2.Context) {
	tpl, err := pongo2.FromFile("templates/" + name)
	if err != nil {
		http.Error(w, "Template not found: "+name, http.StatusInternalServerError)
		return
	}
	out, err := tpl.Execute(data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(out))
}

func Init(params InitParams) (http.Handler, *gorm.DB) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var db *gorm.DB
	if params.DB == nil {
		db, err = initPG()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
	}

	// if err := db.AutoMigrate(&entities.User{}, &entities.Beacon{}, &entities.Media{}, &entities.Depth{}); err != nil {
	// 	log.Fatalf("Failed to auto migrate: %v", err)
	// }
	// db.Debug()

	devMode := params.IsDevMode
	if !params.IsDevMode {
		flag.BoolVar(&devMode, "dev", false, "Run in development mode")
		flag.Parse()
	}

	fmt.Println("Running in dev mode:", devMode)

	var ctx api.BuildContext
	if devMode {
		ctx = esbuild.InitDevMode(params.EsbuildOpts)
		defer ctx.Dispose()
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
					Render(w, tmplName, pongo2.Context{"title": tmplName})
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
			Render(w, "index.twig", pongo2.Context{"title": "Home"})
		})
	}

	for _, setupConfig := range params.RouterSetupFuncs {
		router := setupConfig.Handler(mux, db, devMode)
		mux.Handle(setupConfig.BasePath, router)
	}

	muxWithLogging := middleware.LoggingMiddleware(mux)
	return muxWithLogging, db
}
