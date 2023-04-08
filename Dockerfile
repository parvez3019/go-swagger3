FROM golang:alpine 
WORKDIR /go/src/main
RUN go install github.com/parvez3019/go-swagger3@latest

ENTRYPOINT ["go-swagger3"]
