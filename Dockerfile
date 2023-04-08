# FROM golang:alpine as builder
# WORKDIR /app
# COPY . .
# RUN apk update && apk add upx ca-certificates openssl && update-ca-certificates
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /bin/binary .
# RUN upx -9 /bin/binary

# FROM gcr.io/distroless/static:nonroot
# WORKDIR /app/
# COPY --from=builder /bin/binary /bin/binary
# ENTRYPOINT ["/bin/binary"]



FROM golang:alpine as builder

WORKDIR /work
COPY . .
RUN apk update && apk add upx ca-certificates openssl && update-ca-certificates
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /bin/swagger .
RUN upx -9 /bin/swagger

FROM gcr.io/distroless/static:nonroot

COPY --from=builder bin/swagger /usr/bin/swagger

ENTRYPOINT ["/usr/bin/swagger"]

