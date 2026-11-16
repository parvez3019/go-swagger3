---
title: "Generate OpenAPI 3 specs from Go annotations"
description: "go-swagger3 generates OpenAPI 3.0 specs from swag-style Go comments and serves Swagger UI for Gin, Echo, chi, Fiber, Hertz and net/http. A swaggo/swag alternative with native OpenAPI 3."
---

# go-swagger3

**Generate OpenAPI 3 specs from Go code annotations.** go-swagger3 reads swag-style `@` comments and struct tags,
writes an [OpenAPI 3.0](https://spec.openapis.org/oas/v3.0.3) document (JSON or YAML), and serves it with Swagger UI
from the web framework you already use.

```go
// @Title Get a pet
// @Param id path int true "Pet ID"
// @Success 200 object Pet "The pet"
// @Route /pets/{id} [get]
func GetPet(c *gin.Context) { /* ... */ }
```

```shell
go install github.com/parvez3019/go-swagger3@latest
go-swagger3 --module-path . --output oas.json --schema-without-pkg
```

## Get started with your framework

| Framework | Guide |
|-----------|-------|
| Gin | [OpenAPI 3 with Gin](frameworks/gin.md) |
| Echo | [OpenAPI 3 with Echo](frameworks/echo.md) |
| chi | [OpenAPI 3 with chi](frameworks/chi.md) |
| Fiber | [OpenAPI 3 with Fiber](frameworks/fiber.md) |
| Hertz | [OpenAPI 3 with Hertz](frameworks/hertz.md) |
| gorilla/mux | [OpenAPI 3 with gorilla/mux](frameworks/gorilla-mux.md) |
| net/http | [OpenAPI 3 with net/http](frameworks/net-http.md) |

Buffalo, Flamingo and Atreugo adapters are listed in [Serving Swagger UI](serving-swagger-ui.md#framework-adapters).

## Reference

| Page | Contents |
|------|----------|
| [Install & CLI](cli.md) | `go install`, prebuilt binaries, Docker; flags; `go-swagger3 fmt` |
| [Annotation reference](annotations.md) | Service info, parameters, headers, responses, tags, routes, struct tags, enums, generics |
| [Security](security.md) | HTTP, API key, OpenID Connect and OAuth2 schemes; scopes |
| [Serving Swagger UI](serving-swagger-ui.md) | net/http handler and all framework adapters |

## Coming from swaggo/swag?

- [go-swagger3 vs swaggo/swag vs go-swagger](comparison.md): which tool outputs what.
- [Migrating from swaggo/swag](MIGRATION_FROM_SWAG.md): annotation-by-annotation mapping.

## More

- [AI agent skills](ai-agent-skills.md) for Claude Code and Cursor.
- [Examples](../examples/): runnable net/http, Gin, Echo and chi samples.
- [Releases](https://github.com/parvez3019/go-swagger3/releases) and [changelog](../CHANGELOG.md).
