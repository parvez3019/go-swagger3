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

ADD . /work
WORKDIR /work

RUN apk --no-cache add ca-certificates shared-mime-info mailcap git build-base binutils-gold

RUN mkdir -p bin &&\
  RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/swagger .
RUN upx -9 /bin/swagger

FROM golang:alpine

COPY --from=cross /work/bin/swagger /usr/bin/swagger

ENTRYPOINT ["/usr/bin/swagger"]
CMD ["--help"]
