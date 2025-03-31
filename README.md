# 🚀 Framework

_A minimal yet powerful web framework with builtin support for ssr html pages and SPA microfrontend splits_

**Features**:

- Simple routing & templating (HTML/Go templates)
- Automatic microfrontend support for multiple SPAs
- 1 Go dependency: esbuild
  - Use anything esbuild supports, like TypeScript, JavaScript, JSX, TSX, and CSS
- NO NodeJS process required - no Webpack, Babel, Vite, Rollup, etc.
- Clean `/templates` + `/app` folder structure
- Autoreload when in development mode
- Optional Accompanying JavaScript library for frontend development
  - Server Sent Events handler, Router, state management, and elements
  -

---

## 🚦 Quickstart

### 1. Project Structure

```shell
your-project/
├── templates/          # HTML templates & components
│   ├── base.html       # Required: Base layout
│   ├── entry.html      # Required: html snippet with js and css imports - autogenerated by esbuild on changes
│   ├── index.html      # Required: Default html page
│   ├── other.html      # "/other" route auto-registered
│   └── app.subroute.html # "/app/" subroute auto-registered for the app - an easy way to add a new SPA
├── static # bundled js files place here, you can .gitignore this
├── frontend/
│   └──  src/
│      └── index.ts        # Your frontend entrypoint, customize and codesplit as needed
└── main.go             # Your server
```

2. Get started with all the defaults:

```go
package main

import (
	"net/http"

	"github.com/bencbradshaw/framework"
)

func main() {
	http.ListenAndServe(":2025", framework.Run(nil))
}
```

3. Override the defaults:

```go
package main

import (
	"net/http"
	"os"

	"github.com/bencbradshaw/framework"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	mux := framework.Run(framework.InitParams{
		IsDevMode: false,
		AutoRegisterTemplateRoutes: false,
		EsbuildOpts: api.BuildOptions{
			EntryPoints: []string{"./app/src/my-app.tsx"},
		},
	})
	// add your own routes, the same as you would with the default mux
	mux.Handle("/api/hello-world", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	http.ListenAndServe(":2025", mux)
}
```

### 🚧 In Progress Features

- **Support for Multiple Base Templates**

  - Allow customization of JavaScript or CSS for different templates.
  - Your home page and about page might need different CSS or JS loaded.

- **Image Hosting**

  - Provide built-in support for serving and managing images.
  - This will be simply a folder served with static files

- **Authentication Guard**

  - Offer a middleware mechanism to enable developers to implement custom authentication logic for route handlers.
  - Rather than providing an authentication system, this will allow developers to use their own authentication systems that can tie into the framework.

- **Customizable Templates Directory**

  - Ensure the `/templates` directory path can be configured as needed.

- **Template Naming Conventions**
  - Clearly define and enforce a consistent naming pattern for templates to improve maintainability.
  - `.html` and `subroute.html`, `base` and `entry`. What do they mean and how do they work
