# Install and CLI reference

## Install

**Go toolchain**

```shell
go install github.com/parvez3019/go-swagger3@latest
```

If you get `command not found: go-swagger3`, add your Go bin directory to `PATH`:

```shell
export PATH="$(go env GOPATH)/bin:$PATH"
```

**Prebuilt binary.** Download an archive for Linux, macOS or Windows (amd64/arm64) from the
[latest release](https://github.com/parvez3019/go-swagger3/releases/latest) and put `go-swagger3` on your `PATH`.

**Docker** (linux/amd64 and linux/arm64)

```shell
docker run -t --rm -v $(pwd):/app -w /app parvez3019/go-swagger3:latest --module-path . --output oas.json
```

Check the installed version with `go-swagger3 --version`.

## Generate a spec

Run from the directory that contains your `go.mod`:

```shell
# go.mod and main.go are in the same directory
go-swagger3 --module-path . --output oas.json --schema-without-pkg

# main.go is somewhere else
go-swagger3 --module-path . --main-file-path ./cmd/api/main.go --output oas.json --schema-without-pkg

# YAML instead of JSON
go-swagger3 --module-path . --output oas.yaml --schema-without-pkg --generate-yaml
```

With Docker, pass the same flags after the image name:

```shell
docker run -t --rm -v $(pwd):/app -w /app parvez3019/go-swagger3:latest \
  --module-path . --main-file-path ./cmd/api/main.go --output oas.json --schema-without-pkg
```

## Flags

| Flag | Description |
|------|-------------|
| `--module-path` | Directory of the Go module to scan for `@` annotations |
| `--main-file-path` | File with the service-level annotations (`@Title`, `@Version`, …); defaults to `main.go` in the module path |
| `--handler-path` | Only scan handler comments under this path |
| `--output` | Output file (default `oas.json`) |
| `--generate-yaml` | Write YAML instead of JSON |
| `--schema-without-pkg` | Name schemas without the package prefix (`User` instead of `model.User`) |
| `--exclude dir1,dir2` | Comma-separated paths to skip while walking the module |
| `--quiet` | Only print warnings and errors |
| `--strict` | Treat Go parsing warnings as fatal errors |
| `--debug` | Verbose logs |
| `--version`, `-v` | Print the version |

> [!NOTE]
> Keep handler annotations out of the file passed as `--main-file-path` (by default `main.go`).
> Service-level tags such as `@Title` are read from every comment in that file, so a handler's `@Title` there
> overrides the API title.

## Format annotations

Align `@` annotations in Go source files, like `go fmt` / `swag fmt`:

```shell
go-swagger3 fmt -d ./
```
