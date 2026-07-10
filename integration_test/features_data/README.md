# Feature roadmap integration fixtures

Characterisation suite for **new** go-swagger3 capabilities (Accept/Produce, generics, composition, framework-era schema tags, etc.).

It is intentionally **separate** from the legacy suite in [`../test_data`](../test_data), which must stay FullMatch against its existing goldens and must not be edited to land new features.

| Suite | Purpose | Golden |
|-------|---------|--------|
| `test_data/` | Regression lock for historical behaviour | `expected.json`, `expected_with_pkg.json` |
| `features_data/` (this dir) | Coverage for roadmap features | `spec/expected_features.json` |

Both use the same style: generate OpenAPI JSON, then `jsondiff.FullMatch` against a locked golden (`Test_FeaturesSpec` in [`../features_test.go`](../features_test.go)).

## Layout

```
features_data/
├── go.mod                 # module root scanned by the parser
├── server/main.go         # service-level annotations (@Title, @SecurityScheme, tags, …)
├── handler/features.go    # operation annotations (@Router, @Param, @Success, …)
├── model/features.go      # schemas, components, struct tags
└── spec/
    ├── expected_features.json   # committed golden — do not edit casually
    └── actual_features.json     # written by the test (gitignored)
```

Parser flags used by the test: `--schema-without-pkg` equivalent (`schemaWithoutPkg: true`), module path `features_data`, main file `features_data/server/main.go`.

## How to run

From the repo root (or `integration_test/`):

```bash
go test ./integration_test/ -run Test_FeaturesSpec -count=1
```

Run all integration tests (legacy + features):

```bash
go test ./integration_test/ -count=1
```

## Updating the golden

Only refresh `expected_features.json` when you **intentionally** change generator output for these fixtures (new feature behaviour or a deliberate fix). Do **not** touch `../test_data/spec/expected*.json` for roadmap work.

1. Change annotations under `server/`, `handler/`, or `model/` as needed.
2. Regenerate and inspect:

```bash
cd integration_test
go test -run Test_FeaturesSpec -count=1   # fails if golden is stale; writes actual_features.json
diff -u features_data/spec/expected_features.json features_data/spec/actual_features.json
```

3. If the diff is correct, promote it:

```bash
cp features_data/spec/actual_features.json features_data/spec/expected_features.json
go test ./integration_test/ -run Test_FeaturesSpec -count=1
```

4. Commit the fixture sources **and** the updated golden together.

## What each fixture covers

### Service (`server/main.go`)

| Annotation | What we assert in the golden |
|------------|------------------------------|
| `@Title` / `@Version` / `@Description` / contact / license | `info.*` |
| `@Server` | `servers` |
| `@tag.name` / `@tag.description` | top-level `tags` |
| `@externalDocs.*` | top-level `externalDocs` |
| `@SecurityScheme` (apiKey + http bearer) | `components.securitySchemes` |
| `@Security BearerAuth` | global `security` |

### Operations (`handler/features.go`)

| Handler | Route | Features under test |
|---------|-------|---------------------|
| `CreatePet` | `POST /pets` | `@Accept xml` → request `text/xml`; `@Produce json`; param attrs (`default`, `minimum`/`maximum`, `Enums`, `Format`, `style`, `explode`, `example`); `@Failure 400 $ref:ErrorBody`; per-op `@Security ApiKeyAuth`; `@Tag` |
| `ListPets` | `GET /pets` | Generics `Page[Item]` → schema `Page_Item`; composition `JSONResult{data=Item}` → `JSONResult_data_Item`; `@deprecated`; op `@externalDocs`; `@x-visibility` extension |
| `GetBag` | `GET /bags` | Schema features via `Bag` + `CustomTypes` responses |
| `UploadPet` | `POST /pets/upload` | `@Accept mpfd` → `multipart/form-data`; form + file params; `@Produce json` on response |

Blank import `_ "github.com/parvez3019/go-swagger3/model"` is required so `model.*` type refs resolve (same pattern as legacy `test_data` handlers).

### Models (`model/features.go`)

| Type / tag | Feature |
|------------|---------|
| `@ResponseComponent` / `@RequestBodyComponent` | Reusable `components.responses` / `components.requestBodies` |
| `oneOf` / `anyOf` / `allOf` / `discriminator` on `PetHolder` | Polymorphism field tags |
| `map[string]string` on `Bag.Labels` | `additionalProperties` (not a fake `"key"` property) |
| `*string` / `*int` + `nullable:"false"` | Pointer → `nullable`, with opt-out |
| `swaggertype:"integer"` | Type override |
| `default` / `deprecated:"true"` / `extensions:"x-internal=…"` | Schema defaults, deprecated, vendor extensions |
| `Page[T]` / `JSONResult` | Generics + `{field=Type}` composition (driven from handler `@Success`) |
| `@Enum StatusEnum` | Enum component registration |

## Adding coverage for a new feature

1. Prefer extending this fixture (new handler and/or model fields) over changing `test_data/`.
2. Keep annotations minimal and named so the golden diff is readable.
3. Update `expected_features.json` only after reviewing the full JSON diff.
4. Mention the new case in the tables above.

## Related docs

- [Migration from swag](../../docs/MIGRATION_FROM_SWAG.md)
- Root [README](../../README.md) — Supported Web Frameworks, CLI flags, annotation reference
