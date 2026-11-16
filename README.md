# go-swagger3

**Generate OpenAPI 3 specs from Go code annotations.** A [swaggo/swag](https://github.com/swaggo/swag) alternative
with native OpenAPI 3.0 output, swag-style comments, generics, and built-in Swagger UI for Gin, Echo, chi, Fiber and
more.

[![CI](https://github.com/parvez3019/go-swagger3/actions/workflows/ci.yml/badge.svg)](https://github.com/parvez3019/go-swagger3/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/parvez3019/go-swagger3.svg)](https://pkg.go.dev/github.com/parvez3019/go-swagger3)
[![Release](https://img.shields.io/github/v/release/parvez3019/go-swagger3)](https://github.com/parvez3019/go-swagger3/releases/latest)
[![codecov](https://codecov.io/gh/parvez3019/go-swagger3/branch/main/graph/badge.svg)](https://codecov.io/gh/parvez3019/go-swagger3)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-parvez3019.github.io-00ADD8)](https://parvez3019.github.io/go-swagger3/)

> [!TIP]
> **[Read the docs](https://parvez3019.github.io/go-swagger3/)** ·
> Guides for [Gin](https://parvez3019.github.io/go-swagger3/frameworks/gin/),
> [Echo](https://parvez3019.github.io/go-swagger3/frameworks/echo/),
> [chi](https://parvez3019.github.io/go-swagger3/frameworks/chi/),
> [Fiber](https://parvez3019.github.io/go-swagger3/frameworks/fiber/) ·
> [Migrating from swag](https://parvez3019.github.io/go-swagger3/MIGRATION_FROM_SWAG/)

## Quick start

**1. Install**

```shell
go install github.com/parvez3019/go-swagger3@latest
```

Prebuilt binaries and a Docker image are also available: see [Install](docs/cli.md#install).

**2. Annotate.** Describe the API in `main.go` and each operation on its handler:

```go
// main.go
// @Version 1.0.0
// @Title Pet Store API
// @Server http://localhost:8080 Local
package main
```

```go
// handlers.go
type Pet struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Rex"`
}

// @Title Get a pet
// @Param id path int true "Pet ID"
// @Success 200 object Pet "The pet"
// @Resource pets
// @Route /pets/{id} [get]
func GetPet(c *gin.Context) { /* ... */ }
```

**3. Generate**

```shell
go-swagger3 --module-path . --output oas.json --schema-without-pkg
```

**4. Serve Swagger UI** (Gin shown; [adapters for other frameworks](docs/serving-swagger-ui.md#framework-adapters)):

```go
import (
	_ "embed"

	"github.com/parvez3019/go-swagger3/swagger"
	ginswagger "github.com/parvez3019/go-swagger3/swagger/gin"
)

//go:embed oas.json
var spec []byte

r.Any("/swagger/*any", ginswagger.WrapHandler(spec, swagger.Title("Pet Store API")))
```

Open `http://localhost:8080/swagger/index.html`.

## How it compares

| | go-swagger3 | [swaggo/swag](https://github.com/swaggo/swag) | [go-swagger](https://github.com/go-swagger/go-swagger) |
|---|---|---|---|
| Output | **OpenAPI 3.0** | Swagger 2.0 (v1); OpenAPI 3.1 in v2 release candidates | Swagger 2.0 only |
| Approach | Comments → spec | Comments → spec | `swagger:` comments → spec; also generates code from a spec |
| Annotation style | `@Param`, `@Success`, `@Router`… (swag-like) | `@Param`, `@Success`, `@Router`… | `swagger:route`, `swagger:model`… |
| Generics | Yes (`Page[Pet]`) | Yes | Limited |
| Swagger UI | Built in: net/http, Gin, Echo, chi, mux, Fiber, Hertz, Buffalo, Flamingo, Atreugo | Separate packages (`gin-swagger`, `echo-swagger`, …) | `swagger serve` command |
| Annotation formatter | `go-swagger3 fmt` | `swag fmt` | No |

Already using swag? Most annotations work unchanged: see the [migration guide](docs/MIGRATION_FROM_SWAG.md).

## Features

- **OpenAPI 3.0** output in JSON or YAML.
- **swag-compatible annotations:** `@Param` with `Enums()`, `default()`, `minimum()`…; `@Accept`/`@Produce`;
  `@Router` alias; `@OperationId`; `@deprecated`.
- **Schemas from Go types:** generics, composition (`JSONResult{data=Order}`), maps, pointers as `nullable`,
  `oneOf`/`anyOf`/`allOf`/`discriminator`, enums, field comments as descriptions.
- **Security:** HTTP, API key, OpenID Connect and OAuth2 schemes, applied globally or per operation.
- **Swagger UI adapters** for ten frameworks, each a separate Go module so you only pull in what you use.
- **Annotation formatter:** `go-swagger3 fmt`.

## Documentation

| Page | Contents |
|------|----------|
| [Install & CLI](docs/cli.md) | Install options, flags, Docker, `fmt` |
| [Annotation reference](docs/annotations.md) | Service info, parameters, headers, responses, tags, routes, struct tags, enums |
| [Security](docs/security.md) | Security schemes, OAuth2 flows and scopes |
| [Serving Swagger UI](docs/serving-swagger-ui.md) | net/http handler and framework adapters |
| [Migrating from swag](docs/MIGRATION_FROM_SWAG.md) | Annotation-by-annotation mapping |
| [Examples](examples/) | Runnable net/http, Gin, Echo and chi samples |

## AI agent skills

Copy these into Claude Code or Cursor projects so agents can annotate existing Go APIs and generate OpenAPI 3 specs
with go-swagger3:

| Agent | Source | Copy to |
|-------|--------|---------|
| Claude Code | [agent-skills/claude/SKILL.md](agent-skills/claude/SKILL.md) | `.claude/skills/go-swagger3-docs/SKILL.md` |
| Cursor | [agent-skills/cursor/cursor_rules.mdc](agent-skills/cursor/cursor_rules.mdc) | `.cursor/rules/go-swagger3.mdc` |

## Limitations

- Go modules only.
- Anonymous (unnamed) embedded struct fields have limited support; prefer named types.
- Output is OpenAPI **3.0.0** (3.1 is planned).
- Routes come from annotations (`@Route` / `@Router`); framework routers are not introspected.

## Contributing

Bug reports, feature requests and pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) to get started,
and look for issues labelled [`good first issue`](https://github.com/parvez3019/go-swagger3/labels/good%20first%20issue).
Using go-swagger3 in a project? Tell us in [Show and tell](https://github.com/parvez3019/go-swagger3/discussions/categories/show-and-tell).

## Acknowledgements

go-swagger3 builds on [mikunalpha/goas](https://github.com/mikunalpha/goas),
[yvasiyarov/swagger](https://github.com/yvasiyarov/swagger) and [uudashr/go-module](https://github.com/uudashr/go-module),
and takes feature inspiration from [swaggo/swag](https://github.com/swaggo/swag).

## Star history

[![Star History Chart](https://api.star-history.com/image?repos=parvez3019/go-swagger3&type=date&legend=top-left)](https://www.star-history.com/?repos=parvez3019%2Fgo-swagger3&type=date&legend=top-left)
