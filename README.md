Why Framework

1. Easy setup with no config. Defaults work right out of the box.

```shell
framework init myapp
```

will create:

```go
// main.go
package main

import (
    "github.com/bencbradshaw/framework"
    "net/http"
)

main() {
    mux := framework.Init()
	http.ListenAndServe(":2025", mux);
}
```

directory structure:

```txt
.
├── app
│   └── src/index.ts // default entry point for esbuild
├── entities // default location for GORM database entities
├── routes // default go http handlers
├── templates // twig/html templates - auto register with `name.route.twig`
├── static // git ignored static files - esbuild result goes here
```

- you now can have:
  - html routes served from /templates
  - static files served from /static
  - .js, .ts, .jsx, dev bundled and served from /static/ with autoreload through esbuild
  - GORM entities in /entities added into container for each route

2. Extend and Override defaults

```go
// main.go
package main

import (
    "github.com/bencbradshaw/framework"
    "net/http"
)

main() {
   mux, db := framework.Init({
        routes: [],
        esbuildOptions: {},
        DB: {},
        container: {},
        templateDir: "",
        staticDir: "",
    });

    http.ListenAndServe(":2025", mux);

}
```
