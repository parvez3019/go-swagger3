BINARY_NAME := go-swagger3
COVERAGE_OUT_DIRECTORY := coverage/lcov

.PHONY: all build install tidy fmt vet lint check test cover cover-lcov clean help

all: build

build:
	go build -o $(BINARY_NAME) .

install:
	go install .

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run --issues-exit-code=1

check: test

test:
	go test -cover ./...

cover:
	mkdir -p coverage
	go test -coverpkg=./... -coverprofile coverage.out.tmp ./...
	(cat coverage.out.tmp | grep -v "mocks/" > coverage.out) && rm coverage.out.tmp
	go tool cover -func=coverage.out > coverage/coverage.txt

cover-lcov:
	go get -u github.com/jandelgado/gcov2lcov
	go test -coverpkg=./... -coverprofile coverage.out.tmp ./...
	(cat coverage.out.tmp | grep -v "mocks/" > coverage.out) && rm coverage.out.tmp
	cat coverage.out | gcov2lcov -outfile=coverage.lcov
	mkdir -p $(COVERAGE_OUT_DIRECTORY)
	genhtml coverage.lcov -o $(COVERAGE_OUT_DIRECTORY) && open $(COVERAGE_OUT_DIRECTORY)/index.html

clean:
	rm -f $(BINARY_NAME) coverage.out coverage.out.tmp coverage.lcov
	rm -rf coverage

help:
	@echo "Available targets:"
	@echo "  build      Build the binary"
	@echo "  install    Install the binary to GOPATH/bin"
	@echo "  tidy       Run go mod tidy"
	@echo "  fmt        Format Go source files"
	@echo "  vet        Run go vet"
	@echo "  lint       Run golangci-lint"
	@echo "  test       Run tests with coverage summary"
	@echo "  check      Alias for test"
	@echo "  cover      Generate coverage report"
	@echo "  cover-lcov Generate HTML coverage report via lcov"
	@echo "  clean      Remove build and coverage artifacts"
