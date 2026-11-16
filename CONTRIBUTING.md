# Contributing to go-swagger3

Thanks for helping! Bug reports, docs fixes, new annotations and framework adapters are all welcome.

## Questions and ideas

- **Questions / "how do I…?"**: open a [Discussion](https://github.com/parvez3019/go-swagger3/discussions).
- **Bugs**: open an [issue](https://github.com/parvez3019/go-swagger3/issues/new/choose) with a minimal Go snippet,
  the command you ran, and the spec you expected vs. got.
- **Bigger changes** (new annotations, behaviour changes): open an issue first so we can agree on the design.

Looking for somewhere to start? Try an issue labelled
[`good first issue`](https://github.com/parvez3019/go-swagger3/labels/good%20first%20issue).

## Development setup

Requirements: Go 1.24+.

```shell
git clone https://github.com/parvez3019/go-swagger3.git
cd go-swagger3
make test        # go test -cover ./...
make lint        # golangci-lint (install: https://golangci-lint.run)
make build       # ./go-swagger3
```

The repository is a Go workspace (`go.work`):

| Path | Module |
|------|--------|
| `/` | `github.com/parvez3019/go-swagger3`: CLI, parser, `swagger` UI handler |
| `swagger/<framework>/` | One module per framework adapter (gin, echo, chi, …) |
| `examples/<framework>/` | Runnable examples |

Inside the repo the workspace resolves adapters against your working tree. To test an adapter the way users get it
(against the released root module), run with the workspace off:

```shell
cd swagger/gin && GOWORK=off go test ./...
```

## Tests

- `integration_test/` holds golden-file tests that generate a full spec from sample code and diff it against the
  expected output. If your change intentionally alters output, update the golden files in the same PR and explain why.
- Add unit tests next to the code you change (`parser/...`).
- Run `make cover` for a coverage report.

## Pull requests

1. Fork, create a branch, and keep the change focused.
2. Run `go fmt ./...`, `make lint` and `make test`.
3. If you add or change an annotation, update [docs/annotations.md](docs/annotations.md) and, if relevant,
   [docs/MIGRATION_FROM_SWAG.md](docs/MIGRATION_FROM_SWAG.md).
4. Open the PR against `main` and fill in the template. CI runs tests for the root module, every adapter and example,
   plus lint.

We aim to respond to issues and PRs within two weeks.

## Releases (maintainers)

Root releases are `vX.Y.Z` tags; pushing one runs GoReleaser (binaries, Homebrew cask) and publishes the Docker
image. Adapters are separate modules tagged `swagger/<framework>/vX.Y.Z`. Push tags by name
(`git push origin v1.2.3`), not with `--tags`, and push at most three tags at once, because GitHub doesn't trigger
workflows for larger tag pushes.
