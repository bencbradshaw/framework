Why Framework

- Easily create html pages for SEO optimized "public" facing pages
- AND create one or multiple JavaScript SPA without running any NodeJS process
  - Utilizing esbuild's Go module, there is no need for Webpack, Babel, Vite, Rollup, or any separate process apart from your Go server. TypeScript, JavaScript, JSX, TSX, and CSS
- Accompanying JavaScript library for frontend development
  - Server-sent events for auto-reload while developing
  - Router, state management, and elements
- **Twig templates** for HTML rendering
- **Native Go HTTP module**

create your project

```txt
├── app
│   └── src
│        └── index.ts // this is the default entry point for your frontend application
├── templates // twig/html templates - auto register with `name.route.twig`
│   ├── index.twig // "/" default route
│   └── about.route.twig // "/about" route
├── main.go // your go application - see below for example
```

- you now can have:
  - html routes served from /templates
  - frontend javascript bundled and served

2. Extend and Override defaults

```go
package main

import (
	"net/http"
	"github.com/bencbradshaw/go-web-framework"
	"github.com/evanw/esbuild/pkg/api"
)


func main() {
	mux := framework.Run()
	// add any other routes needed for you application, e.g.
	// mux.Handle("/api/chat", handlers.HandleChatRequest)
	print("Server started at http://localhost:2025 \n")
	http.ListenAndServe(":2025", mux)
}

```
