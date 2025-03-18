Why Go Web Framework

- server side rendering and client side javascript apps
- accompanying javascript library for frontend development
  - server sent events for auto-reload while developing
  - router, state, elements
- esbuild for fast and simple ts, js, jsx, tsx, css support for bundling, minifying, code splitting
- twig templates for html rendering
- native go http module

create your project

```txt
├── app
│   └── src
│        └── index.ts // default entry point for esbuild
├── templates // twig/html templates - auto register with `name.route.twig`
│   ├── index.twig // "/" default route
│   └── about.twig // "/about" route
├── main.go // your go application - see below for example
```

- you now can have:
  - html routes served from /templates
  - frontend javascript bundled and served

2. Extend and Override defaults

```go
// main.go
package main

import (
	"net/http"
	"os"
	"github.com/bencbradshaw/go-web-framework"
	"github.com/evanw/esbuild/pkg/api"
)


func main() {
	if os.Getenv("BUILD") == "true" {
		buildParams := framework.InitParams{
			EsbuildOpts: api.BuildOptions{
				EntryPoints:       []string{"./app/src/index.ts"},
				MinifyWhitespace:  true,
				MinifyIdentifiers: true,
				MinifySyntax:      true,
				Sourcemap:         api.SourceMapNone,
			},
			AutoRegisterTemplateRoutes: true,
		}
		framework.Build(buildParams)
		print("Build complete \n")
		return
	}
	mux := framework.Run(framework.InitParams{
		IsDevMode: true,
		EsbuildOpts: api.BuildOptions{
			EntryPoints: []string{"./app/src/index.ts"},
		},
		AutoRegisterTemplateRoutes: true,
	})
	// add any other routes needed for you application, e.g.
	// mux.Handle("/api/chat", handlers.HandleChatRequest)
	print("Server started at http://localhost:2025 \n")
	http.ListenAndServe(":2025", mux)
}

```
