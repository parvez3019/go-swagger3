# Migrating from swaggo/swag to go-swagger3

go-swagger3 generates **OpenAPI 3.0** natively. swag (v1) generates Swagger 2.0; swag v2 targets OpenAPI 3.1 (RC). This guide maps common swag annotations and UI mounts to go-swagger3.

## Quick differences

| Topic | swag | go-swagger3 |
|-------|------|-------------|
| Spec version | Swagger 2.0 (`swagger: "2.0"`) | OpenAPI 3.0.0 (`openapi: "3.0.0"`) |
| Host / base path | `@host`, `@BasePath` | `@Server {url} {description}` (repeatable) |
| Content types | `@Accept` / `@Produce` | Same (`@Accept`, `@Produce`) |
| Security apply | Per-operation `@Security` | Global `@Security` **and** per-operation `@Security` |
| Security define | `@securityDefinitions.*` | `@SecurityScheme` / `@SecurityScope` |
| UI middleware | Separate repos (`gin-swagger`, …) | In-repo adapters under `swagger/<framework>/` |
| Format comments | `swag fmt` | `go-swagger3 fmt -d ./` |
| Generics | `Foo[Bar]` | `Foo[Bar]` (schema id `Foo_Bar`) |
| Composition | `JSONResult{data=Order}` | Same `{field=Type}` sugar |

## Annotation mapping

### General API info

| swag | go-swagger3 |
|------|-------------|
| `@title` | `@Title` |
| `@version` | `@Version` |
| `@description` | `@Description` |
| `@termsOfService` | `@TermsOfServiceUrl` |
| `@contact.name` / `.email` / `.url` | `@ContactName` / `@ContactEmail` / `@ContactURL` |
| `@license.name` / `.url` | `@LicenseName` / `@LicenseURL` |
| `@host` + `@BasePath` | `@Server https://example.com/api/v1 Production` |
| `@tag.name` / `@tag.description` | `@tag.name` / `@tag.description` |
| `@externalDocs.description` / `.url` | Same |

### Operations

| swag | go-swagger3 |
|------|-------------|
| `@Summary` | `@Title` (summary) |
| `@Description` | `@Description` |
| `@Tags` | `@Resource` or `@Tag` |
| `@Accept` / `@Produce` | Same |
| `@Param` | Same shape; also supports `Enums()`, `default()`, `minimum()`, `Format()`, `style()`, `explode()` |
| `@Success` / `@Failure` / `@Response` | `@Success` / `@Failure` |
| `@Header` (response) | `@ResponseHeader` |
| `@Router /path [get]` | `@Route` or `@Router` |
| `@id` | `@OperationId` |
| `@deprecated` | `@deprecated` |
| `@Security ApiKeyAuth` | `@Security ApiKeyAuth` (on handler or globally) |

### Struct tags

| swag | go-swagger3 |
|------|-------------|
| `example`, `enums`, validation tags | Same family (`example`, `enum`, `minimum`, …) |
| `swaggertype` | `swaggertype` or `go-swagger3:"type=…"` |
| `swaggerignore:"true"` | `skip:"true"` or `go-swagger3:"-"` / `json:"-"` |
| Field comments as descriptions | Supported (and `description` tag) |

## Generate

```bash
# swag
swag init -g cmd/api/main.go

# go-swagger3
go-swagger3 --module-path . --main-file-path ./cmd/api/main.go --output oas.json --schema-without-pkg
```

Useful flags: `--exclude`, `--quiet`, `--handler-path`, `--strict`, `--generate-yaml`, `--debug`.

Format annotations:

```bash
go-swagger3 fmt -d ./
```

## Serve Swagger UI

### net/http

```go
import "github.com/parvez3019/go-swagger3/swagger"

http.Handle("/swagger/", swagger.Handler(spec, swagger.Title("My API")))
```

### gin

```go
import ginswagger "github.com/parvez3019/go-swagger3/swagger/gin"

r.Any("/swagger/*any", ginswagger.WrapHandler(spec))
```

```bash
go get github.com/parvez3019/go-swagger3/swagger/gin
```

### echo

```go
import echoswagger "github.com/parvez3019/go-swagger3/swagger/echo"

e.Any("/swagger/*", echoswagger.WrapHandler(spec))
```

### chi

```go
import chiswagger "github.com/parvez3019/go-swagger3/swagger/chi"

r.Handle("/swagger/*", chiswagger.WrapHandler(spec))
```

### gorilla/mux

```go
import muxswagger "github.com/parvez3019/go-swagger3/swagger/mux"

r.PathPrefix("/swagger/").Handler(muxswagger.WrapHandler(spec))
```

### fiber

```go
import fiberswagger "github.com/parvez3019/go-swagger3/swagger/fiber"

app.All("/swagger/*", fiberswagger.WrapHandler(spec))
```

### hertz / buffalo / flamingo / atreugo

Same pattern — import `github.com/parvez3019/go-swagger3/swagger/<framework>` and call `WrapHandler(spec)`. See [examples/](../examples/) and each adapter’s package comment.

Each adapter is a **separate Go module** so you only pull the frameworks you use.

## Security schemes

```go
// swag
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// go-swagger3
// @SecurityScheme ApiKeyAuth apiKey header Authorization
// @Security ApiKeyAuth
```

HTTP bearer / OAuth2 / OpenID Connect use `@SecurityScheme` forms documented in the README.

## What stays annotation-driven

Like swag, go-swagger3 does **not** introspect gin/echo/fiber route tables. Handlers still need `@Router` (or `@Route`) comments.

## OpenAPI 3.1

go-swagger3 targets OpenAPI **3.0.0** today. A 3.1 output mode is planned once 3.0 coverage is complete; prefer 3.0 tooling (Swagger UI, spectral, openapi-generator) for now.
