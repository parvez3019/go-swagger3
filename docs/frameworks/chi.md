---
title: "OpenAPI 3 (Swagger) with chi"
description: "Generate an OpenAPI 3 spec from Go annotations in a chi app and serve Swagger UI with go-swagger3."
---

# OpenAPI 3 (Swagger) with chi

Generate an **OpenAPI 3.0** spec from comments in a [chi](https://github.com/go-chi/chi) application and serve it with
Swagger UI at `/swagger/index.html`. go-swagger3 reads swag-style `@` annotations, so if you use swaggo/swag with
http-swagger (`swaggo/http-swagger`) today, most of your comments work unchanged.

## 1. Install

```shell
go install github.com/parvez3019/go-swagger3@latest
go get github.com/go-chi/chi/v5 github.com/parvez3019/go-swagger3/swagger/chi
```

The [`swagger/chi`](https://pkg.go.dev/github.com/parvez3019/go-swagger3/swagger/chi) adapter is a separate Go module, so your project only pulls in chi, not every framework go-swagger3 supports.

## 2. Annotate

Describe the API in `main.go` and each operation on its handler. Keep handlers in their own file: service-level
annotations such as `@Title` are read from every comment in `main.go`.

```go
// handlers.go
package main

import (
	"encoding/json"
	"net/http"
)

type Pet struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Rex"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"pet not found"`
}

// GetPet returns a pet by ID.
// @Title Get a pet
// @Description Returns a single pet.
// @Param id path int true "Pet ID" "1"
// @Success 200 object Pet "The pet"
// @Failure 404 object ErrorResponse "Pet not found"
// @Resource pets
// @Route /pets/{id} [get]
func GetPet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Pet{ID: 1, Name: "Rex"})
}
```

## 3. Generate the spec

```shell
go-swagger3 --module-path . --output oas.json --schema-without-pkg
```

This writes `oas.json` (add `--generate-yaml` and `--output oas.yaml` for YAML). Re-run it whenever annotations
change, for example from a `//go:generate` directive or a Makefile target.

## 4. Serve Swagger UI

```go
// main.go
// @Version 1.0.0
// @Title Pet Store API
// @Description A tiny example API built with chi.
// @Server http://localhost:8080 Local
package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/parvez3019/go-swagger3/swagger"
	chiswagger "github.com/parvez3019/go-swagger3/swagger/chi"
)

//go:embed oas.json
var spec []byte

func main() {
	r := chi.NewRouter()
	r.Get("/pets/{id}", GetPet)
	r.Handle("/swagger/*", chiswagger.WrapHandler(spec, swagger.Title("Pet Store API")))
	log.Fatal(http.ListenAndServe(":8080", r))
}
```

Run `go run .` and open [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html). The raw
spec is served at `/swagger/doc.json`.

## Next steps

- [Annotation reference](../annotations.md): parameters, headers, responses, enums, generics.
- [Security](../security.md): bearer tokens, API keys, OAuth2.
- [Migrating from swaggo/swag](../MIGRATION_FROM_SWAG.md): annotation-by-annotation mapping.
- [Other frameworks](../serving-swagger-ui.md#framework-adapters).
