- copy the .env.template and create a .env file
- add db credentials to .env file

* install dependencies

```shell
go mod download
```

- run the app

```shell
make
```

Why Framework

1. Easy setup with no config. Defaults work right out of the box.

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

- you now can have:
  - html templates auto loaded from the templates folder
  - static files served from /static
  - .js, .ts, .jsx, dev bundled and served from /static/ with autoreload through esbuild

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
