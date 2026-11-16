# Annotation reference

go-swagger3 reads `@` annotations from Go comments and struct tags. Service-level annotations go in your main file;
operation annotations go on handler functions; schemas come from your Go types.

- [Service description](#service-description)
- [Operations (handler functions)](#operations)
  - [Title and description](#title-and-description)
  - [Parameters](#parameters)
  - [Request headers](#request-headers)
  - [Responses](#responses)
  - [Response headers](#response-headers)
  - [Tags (resource)](#tags)
  - [Route](#route)
  - [Content types, operation ID, deprecation](#content-types-operation-id-deprecation)
- [Schemas (struct tags)](#schemas)
  - [Enums](#enums)
  - [Generics and composition](#generics-and-composition)
- [Security](security.md)

Coming from swaggo/swag? The [migration guide](MIGRATION_FROM_SWAG.md) maps every swag annotation to its
go-swagger3 equivalent.

## Service description

Put these in the file passed as `--main-file-path` (by default `main.go` in the module path). They describe the API
as a whole.

```go
// @Version 1.0.0
// @Title Backend API
// @Description API usually works as expected. But sometimes its not true.
// @ContactName Parvez
// @ContactEmail abce@email.com
// @ContactURL http://someurl.oxox
// @TermsOfServiceUrl http://someurl.oxox
// @LicenseName MIT
// @LicenseURL https://en.wikipedia.org/wiki/MIT_License
// @Server http://www.fake.com Server-1
// @Server http://www.fake2.com Server-2
// @Security AuthorizationHeader read write
// @SecurityScheme AuthorizationHeader http bearer Input your token
package main
```

> [!NOTE]
> Keep handler annotations out of this file. Service-level tags are read from every comment in it, so a handler's
> `@Title` there overrides the API title.

## Operations

Annotate each handler's godoc to document its route, inputs and outputs.

```go
type User struct {
	ID   uint64 `json:"id" example:"100" description:"User identity"`
	Name string `json:"name" example:"Parvez"`
}

type UsersResponse struct {
	Data []User `json:"users" example:"[{\"id\":100, \"name\":\"Parvez\"}]"`
}

type Error struct {
	Code string `json:"code"`
	Msg  string `json:"msg" skip:"true"` // skip:"true" leaves the field out of the spec
}

type ErrorResponse struct {
	ErrorInfo Error `json:"error"`
}

// RequestHeaders represents the model for header params
// @HeaderParameters RequestHeaders
type RequestHeaders struct {
	Authorization string `json:"Authorization"`
}

// @Title Get user list of a group.
// @Description Get users related to a specific group.
// @Header model.RequestHeaders
// @Param  groupID  path  int  true  "Id of a specific group." "120"
// @Success  200  object  UsersResponse  "UsersResponse JSON"
// @Failure  400  object  ErrorResponse  "ErrorResponse JSON"
// @Resource users
// @Route /api/group/{groupID}/users [get]
func GetGroupUsers() {
	// ...
}

// @Title Create a user.
// @Description Create a new user.
// @Param  user  body  User  true  "Info of a user."
// @Success  200  object  User           "UsersResponse JSON"
// @Failure  400  object  ErrorResponse  "ErrorResponse JSON"
// @Resource users
// @Route /api/user [post]
func PostUser() {
	// ...
}
```

### Title and description

```
@Title {title}
@Title Get user list of a group.

@Description {description}
@Description Get users related to a specific group.
```

- `{title}`: the operation summary.
- `{description}`: the operation description.

### Parameters

```
@Param  {name}   {in}  {goType}  {required}  {description}             {example}
@Param  groupID  path  int       true        "Id of a specific group." "120"
```

- `{name}`: the parameter name.
- `{in}`: `path`, `query`, `form`, `header`, `cookie`, `body` or `file`.
- `{goType}`: the Go type. Ignored when `{in}` is `file`.
- `{required}`: `true`, `false`, `required` or `optional`.
- `{description}`: must be quoted.
- `{example}`: **optional** example value. Must be quoted.

Constraints and defaults can follow the description, swag-style:

```
@Param status query string false "Filter by status" Enums(available,sold) default(available)
@Param limit  query int    false "Page size"        minimum(1) maximum(100)
```

Supported attributes: `Enums()`, `default()`, `minimum()`, `maximum()`, `minlength()`, `maxlength()`, `Format()`,
`collectionFormat()`, `style()`, `explode()`, `example()`.

To override the example of an object, use the `override-example` struct tag:

```go
type Request struct {
	Version model.Version `json:"version" override-example:"11.0.0"`
}
```

### Request headers

Document request headers from a struct, either inline on an operation or as a reusable component:

```
@Header            {goType}
@Header            model.RequestHeaders

@HeaderParameters  {goType}
@HeaderParameters  RequestHeaders
```

- `@Header` on a handler adds each field of the type as a header parameter of that operation.
- `@HeaderParameters` on a type registers its fields as parameters in the `components` section for reference.

### Responses

```
@Success  {status}  {jsonType}  {goType}       {description}
@Success  200       object      UsersResponse  "UsersResponse JSON"

@Failure  {status}  {jsonType}  {goType}       {description}
@Failure  400       object      ErrorResponse  "ErrorResponse JSON"
```

- `{status}`: the HTTP status code.
- `{jsonType}`: `object` or `array`.
- `{goType}`: the Go type.
- `{description}`: must be quoted.

### Response headers

```
@ResponseHeader  {status}  {headerName}  {type}  "{description}"     "{example}"
@ResponseHeader  200       Set-Cookie    string  "JWT token cookie"  "accessToken=bzRslFay1i...; Path=/; HttpOnly; Secure; SameSite=Lax"
```

- `{status}`: the status code of the response this header belongs to.
- `{headerName}`: e.g. `Set-Cookie`, `X-Request-Id`.
- `{type}`: `string`, `integer`, `boolean` or `number`.
- `{description}`: must be quoted.
- `{example}`: **optional**, must be quoted.

`@ResponseHeader` can come before or after the matching `@Success`/`@Failure`; the response is created if it doesn't
exist yet. Multiple headers per status are supported:

```go
// @Success        200  object  TokenResponse  "Login successful"
// @ResponseHeader 200  Set-Cookie    string  "Access token cookie"  "accessToken=bzRslFay1i...; Path=/; HttpOnly; Secure; SameSite=Lax"
// @ResponseHeader 200  X-Request-Id  string  "Unique request identifier"
func Login() {
	// ...
}
```

Output:

```json
"200": {
  "description": "Login successful",
  "headers": {
    "Set-Cookie": {
      "description": "Access token cookie",
      "schema": {
        "type": "string",
        "example": "accessToken=bzRslFay1i...; Path=/; HttpOnly; Secure; SameSite=Lax"
      }
    },
    "X-Request-Id": {
      "description": "Unique request identifier",
      "schema": { "type": "string" }
    }
  },
  "content": {
    "application/json": {
      "schema": { "$ref": "#/components/schemas/TokenResponse" }
    }
  }
}
```

### Tags

```
@Resource {resource}
@Resource users

@Tag {tag}
@Tag users
```

Both set the operation's tag. Describe tags at service level with `@tag.name`, `@tag.description`,
`@tag.docs.url` and `@tag.docs.description`.

### Route

```
@Route {path}           {method}
@Route /api/user        [post]
@Route /api/user/{id}   [get]
```

- `{path}`: the URL path, with `{param}` placeholders.
- `{method}`: the HTTP method, in brackets.

`@Router` is accepted as an alias, so swag annotations work unchanged.

### Content types, operation ID, deprecation

```
@Accept      json
@Produce     json
@OperationId ListPets
@deprecated
@externalDocs.description More about pets
@externalDocs.url         https://example.com/docs/pets
```

- `@Accept` / `@Produce` set request and response content types. MIME aliases such as `json`, `xml` and `mpfd` work.
- `@OperationId` must be unique across the module.

## Schemas

Schemas are generated from your Go types. Add OpenAPI keywords as struct tags:

```go
type Filter struct {
	Rating       int    `json:"rating"`
	Type         string `json:"type"`
	Distance     int64  `json:"distance" minimum:"1" maximum:"100"`
	DistrictCode string `json:"district_code" pattern:"[a-z]{2}\\d[a-z]\\s\\d[a-z]{2}"`
}
```

Supported tags:

| Tag | Notes |
|-----|-------|
| `json` | Field name; `json:"-"` skips the field |
| `description`, `title`, `example`, `default` | Field comments are also used as descriptions |
| `type`, `format`, `swaggertype` | Override the inferred type; also `go-swagger3:"type=…"` |
| `required` | |
| `enum`, `$ref` | See [Enums](#enums) |
| `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum` | Numbers |
| `minLength`, `maxLength`, `pattern` | Strings |
| `minItems`, `maxItems`, `uniqueItems` | Arrays |
| `minProperties`, `maxProperties`, `additionalProperties` | Objects; maps become `additionalProperties` |
| `nullable`, `readOnly`, `writeOnly`, `deprecated` | Booleans; pointer fields are `nullable` |
| `oneOf`, `anyOf`, `allOf`, `discriminator` | Polymorphic schemas |
| `skip` | `skip:"true"` leaves the field out; also `go-swagger3:"-"` |

### Enums

Declare an enum as a struct with a single field named after the struct, then reference it with `$ref`:

```go
// @Enum CountriesEnum
type CountriesEnum struct {
	CountriesEnum string `enum:"india,china,mexico,japan" example:"india"`
}

type Request struct {
	Name    string `json:"name" example:"Parvez"`
	Country string `json:"country" $ref:"CountriesEnum"`
}
```

### Generics and composition

- **Generics:** `@Success 200 object Page[Pet] "A page of pets"` generates a `Page_Pet` schema.
- **Composition:** `@Success 200 object JSONResult{data=Order} "Order"` replaces the `data` field of `JSONResult`
  with `Order`, the same sugar swag uses.
