# go-swagger3 documentation

go-swagger3 generates [OpenAPI 3](https://spec.openapis.org/oas/v3.0.3) specs from Go code annotations and struct
tags, and serves them with Swagger UI from your existing web framework.

| Page | Contents |
|------|----------|
| [Install & CLI](cli.md) | `go install`, prebuilt binaries, Docker; flags; `go-swagger3 fmt` |
| [Annotation reference](annotations.md) | Service info, parameters, headers, responses, tags, routes, struct tags, enums, generics |
| [Security](security.md) | HTTP, API key, OpenID Connect and OAuth2 schemes; scopes |
| [Serving Swagger UI](serving-swagger-ui.md) | net/http handler; Gin, Echo, chi, mux, Fiber, Hertz, Buffalo, Flamingo, Atreugo adapters |
| [Migrating from swaggo/swag](MIGRATION_FROM_SWAG.md) | Annotation-by-annotation mapping from swag |
| [Examples](../examples/) | Runnable net/http, Gin, Echo and chi samples |

New here? Start with the [quick start](../README.md#quick-start).
