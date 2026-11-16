---
title: "go-swagger3 vs swaggo/swag vs go-swagger"
description: "Compare Go OpenAPI generators. go-swagger3 outputs OpenAPI 3.0 from swag-style annotations; swaggo/swag outputs Swagger 2.0; go-swagger supports Swagger 2.0 only."
---

# go-swagger3 vs swaggo/swag vs go-swagger

Three Go tools generate API specs from source code comments. The biggest difference is the **spec version** they
produce.

| | go-swagger3 | [swaggo/swag](https://github.com/swaggo/swag) | [go-swagger](https://github.com/go-swagger/go-swagger) |
|---|---|---|---|
| Output | **OpenAPI 3.0** | Swagger 2.0 (v1); OpenAPI 3.1 in v2 release candidates | Swagger 2.0 only |
| Approach | Comments → spec | Comments → spec | `swagger:` comments → spec; also generates server/client code from a spec |
| Annotation style | `@Param`, `@Success`, `@Router`… | `@Param`, `@Success`, `@Router`… | `swagger:route`, `swagger:model`, `swagger:parameters`… |
| Generics | Yes (`Page[Pet]`) | Yes | Limited |
| Swagger UI | Built in for net/http, Gin, Echo, chi, mux, Fiber, Hertz, Buffalo, Flamingo, Atreugo | Separate packages (`gin-swagger`, `echo-swagger`, …) | `swagger serve` command |
| Annotation formatter | `go-swagger3 fmt` | `swag fmt` | No |
| Output formats | JSON, YAML | JSON, YAML, Go (`docs.go`) | JSON, YAML |

*Checked against swag v1.16.6 / v2.0.0-rc6 and go-swagger v0.36.7 (November 2026).*

## When to choose which

**Choose go-swagger3 if** you want OpenAPI 3 today from annotations that look like swag's. Most swag comments work
unchanged, so moving an existing project is mostly a matter of running a different generator. OpenAPI 3 brings
`oneOf`/`anyOf`, `nullable`, multiple servers, request bodies with several content types, and better support in
modern tooling (client generators, API gateways, Postman, Redocly).

**Choose swaggo/swag if** Swagger 2.0 is enough for your consumers, or you need its generated `docs.go` package. Its
v2 line targets OpenAPI 3.1 but is still a release candidate.

**Choose go-swagger if** you want to **generate** Go servers and clients from a hand-written Swagger 2.0 spec
(design-first), rather than generate a spec from code.

## Moving from swag

1. Install go-swagger3: `go install github.com/parvez3019/go-swagger3@latest`.
2. Run it on your module: `go-swagger3 --module-path . --output oas.json --schema-without-pkg`.
3. Replace `@host`/`@BasePath` with `@Server`, and `@securityDefinitions.*` with `@SecurityScheme`.
4. Swap `gin-swagger`/`echo-swagger` for the go-swagger3 adapter for your framework, e.g.
   [Gin](frameworks/gin.md) or [Echo](frameworks/echo.md).

The [migration guide](MIGRATION_FROM_SWAG.md) maps every swag annotation to its go-swagger3 equivalent.
