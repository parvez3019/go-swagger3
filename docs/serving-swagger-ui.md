---
title: "Serve Swagger UI in Go"
description: "Serve Swagger UI for an OpenAPI 3 spec from net/http, Gin, Echo, chi, gorilla/mux, Fiber, Hertz, Buffalo, Flamingo or Atreugo."
---

# Serving Swagger UI

Serve the generated OpenAPI file with an interactive UI at `http://localhost:8080/swagger/index.html`
(the same idea as [swaggo/swag](https://github.com/swaggo/swag)'s `gin-swagger`, but built in).

| Path | Content |
|------|---------|
| `/swagger/index.html` | Swagger UI |
| `/swagger/doc.json` | OpenAPI 3 document |

## net/http

```shell
go get github.com/parvez3019/go-swagger3/swagger
```

```go
package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/parvez3019/go-swagger3/swagger"
)

//go:embed oas.json
var spec []byte

func main() {
	http.Handle("/swagger/", swagger.Handler(spec, swagger.Title("My API")))
	// your API routes...
	log.Println("Swagger UI at http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

To read the spec from disk at startup instead of embedding it:

```go
h, err := swagger.HandlerFromFile("oas.json")
if err != nil {
	log.Fatal(err)
}
http.Handle("/swagger/", h)
```

## Framework adapters

Thin adapters wrap the core `swagger.Handler`. Each adapter is a **separate Go module**, so you only pull in the
framework you use:

```shell
go get github.com/parvez3019/go-swagger3/swagger/gin
```

| Framework | Import | Mount |
|-----------|--------|-------|
| [net/http](https://pkg.go.dev/net/http) | `github.com/parvez3019/go-swagger3/swagger` | `http.Handle("/swagger/", swagger.Handler(spec))` |
| [Gin](https://github.com/gin-gonic/gin) | `.../swagger/gin` | `r.Any("/swagger/*any", ginswagger.WrapHandler(spec))` |
| [Echo](https://github.com/labstack/echo) | `.../swagger/echo` | `e.Any("/swagger/*", echoswagger.WrapHandler(spec))` |
| [gorilla/mux](https://github.com/gorilla/mux) | `.../swagger/mux` | `r.PathPrefix("/swagger/").Handler(muxswagger.WrapHandler(spec))` |
| [chi](https://github.com/go-chi/chi) | `.../swagger/chi` | `r.Handle("/swagger/*", chiswagger.WrapHandler(spec))` |
| [Fiber](https://github.com/gofiber/fiber) | `.../swagger/fiber` | `app.All("/swagger/*", fiberswagger.WrapHandler(spec))` |
| [Hertz](https://github.com/cloudwego/hertz) | `.../swagger/hertz` | see [package docs](https://pkg.go.dev/github.com/parvez3019/go-swagger3/swagger/hertz) |
| [Buffalo](https://github.com/gobuffalo/buffalo) | `.../swagger/buffalo` | see [package docs](https://pkg.go.dev/github.com/parvez3019/go-swagger3/swagger/buffalo) |
| [Flamingo](https://github.com/i-love-flamingo/flamingo) | `.../swagger/flamingo` | `http.Handler` wrapper |
| [Atreugo](https://github.com/savsgio/atreugo) | `.../swagger/atreugo` | see [package docs](https://pkg.go.dev/github.com/parvez3019/go-swagger3/swagger/atreugo) |

Gin example:

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/parvez3019/go-swagger3/swagger"
	ginswagger "github.com/parvez3019/go-swagger3/swagger/gin"
)

r := gin.Default()
r.Any("/swagger/*any", ginswagger.WrapHandler(spec, swagger.Title("My API")))
```

Runnable samples for net/http, Gin, Echo and chi are in [examples/](../examples/).
