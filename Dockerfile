# Build stage runs on the builder's native platform and cross-compiles for the
# target, so multi-arch builds don't need QEMU emulation.
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
# Build the checked-out source (the tagged release), not @latest from the proxy.
RUN CGO_ENABLED=0 GOWORK=off GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags "-s -w -X github.com/parvez3019/go-swagger3/app.version=${VERSION}" \
    -o /out/go-swagger3 .

# Keep a Go toolchain in the runtime image: go-swagger3 resolves modules and
# GOPATH of the project it parses.
FROM golang:1.24-alpine
WORKDIR /go/src/main
COPY --from=build /out/go-swagger3 /usr/local/bin/go-swagger3

ENTRYPOINT ["go-swagger3"]
