---
title: "OpenAPI 3 security schemes from Go annotations"
description: "Define bearer, basic, API key, OpenID Connect and OAuth2 security schemes and scopes in Go comments with go-swagger3."
---

# Security

If authorization is required, define security schemes with `@SecurityScheme [name] [type] [parameters]` and apply
them with `@Security [scheme-name] [scope1] [scope2] [...]`.

All examples here use `MyApiAuth` as the name. The name can be anything; multiple named schemes are supported. Each
scheme needs its own name, except OAuth2, which supports multiple flows under the same name.

## Scheme types

| Type | Description | Parameters | Example |
|------|-------------|------------|---------|
| HTTP | HTTP authentication using the `Authorization` header | scheme: any [HTTP authentication scheme](https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml) | `@SecurityScheme MyApiAuth basic` |
| APIKey | An API key passed with the request | in: `header`, `query` or `cookie`; name: the field that carries the key | `@SecurityScheme MyApiAuth apiKey header X-MyCustomHeader` |
| OpenIdConnect | Delegate to an OpenID Connect provider | url: the provider's discovery URL | `@SecurityScheme MyApiAuth openIdConnect https://example.com/.well-known/openid-configuration` |
| OAuth2AuthCode | OAuth2 authorization code flow | authorizationUrl, tokenUrl | `@SecurityScheme MyApiAuth oauth2AuthCode /oauth/authorize /oauth/token` |
| OAuth2Implicit | OAuth2 implicit flow | authorizationUrl | `@SecurityScheme MyApiAuth oauth2Implicit /oauth/authorize` |
| OAuth2ResourceOwnerCredentials | OAuth2 resource owner password flow | tokenUrl | `@SecurityScheme MyApiAuth oauth2ResourceOwnerCredentials /oauth/token` |
| OAuth2ClientCredentials | OAuth2 client credentials flow | tokenUrl | `@SecurityScheme MyApiAuth oauth2ClientCredentials /oauth/token` |

Any text after the last parameter becomes the description, for example
`@SecurityScheme MyApiAuth basic Login with your admin credentials`.

## Applying schemes

Apply **global** security in the service description and/or **per-operation** security on a handler:

```go
// Global (service description)
// @Security MyApiAuth read_user write_user

// Per-operation (handler godoc)
// @Security MyApiAuth read_user
```

## OAuth2 scopes

Define scopes with `@SecurityScope [scheme-name] [scope-code] [scope-description]`:

```go
// @SecurityScope MyApiAuth read_user Read a user from the system
// @SecurityScope MyApiAuth write_user Write a user to the system
```
