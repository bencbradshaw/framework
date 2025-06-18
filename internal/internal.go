package internal

import (
	"fmt"
	"net/http"

	"github.com/evanw/esbuild/pkg/api"
)

type InitParams struct {
	Mux                        *http.ServeMux
	IsDevMode                  bool
	EsbuildOpts                api.BuildOptions
	AutoRegisterTemplateRoutes bool
	AuthGuard                  func(http.Handler) http.Handler
	TemplateDir                string
	StaticDir                  string
}

func MergeDefaults(providedInitParams InitParams) InitParams {
	fmt.Println("Merging default options for internal use")
	initParams := InitParams{
		IsDevMode: true,
		EsbuildOpts: api.BuildOptions{
			EntryPoints: []string{"./frontend/src/index.ts"},
		},
		AutoRegisterTemplateRoutes: true,
		AuthGuard:                  nil,
		TemplateDir:                "templates",
		StaticDir:                  "static",
	}
	if !providedInitParams.IsDevMode {
		initParams.IsDevMode = providedInitParams.IsDevMode
	}
	if len(providedInitParams.EsbuildOpts.EntryPoints) > 0 {
		initParams.EsbuildOpts = providedInitParams.EsbuildOpts
	}
	if !providedInitParams.AutoRegisterTemplateRoutes {
		initParams.AutoRegisterTemplateRoutes = providedInitParams.AutoRegisterTemplateRoutes
	}
	if providedInitParams.AuthGuard != nil {
		initParams.AuthGuard = providedInitParams.AuthGuard
	}
	if providedInitParams.TemplateDir != "templates" {
		initParams.TemplateDir = providedInitParams.TemplateDir
	}
	if providedInitParams.StaticDir != "static" {
		initParams.StaticDir = providedInitParams.StaticDir
	}
	return initParams

}
