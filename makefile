COVERAGE_OUT_DIRECTORY="coverage/lcov"

.PHONY: check
check:
	go test -cover ./...

.PHONY: cover
cover:
	go test -coverpkg=./... -coverprofile coverage.out.tmp ./...
	(cat coverage.out.tmp | grep -v "mocks/" > coverage.out) && rm coverage.out.tmp
	go tool cover -func=coverage.out > coverage/coverage.txt

.PHONY: cover
cover-lcov:
	go get -u github.com/jandelgado/gcov2lcov
	go test -coverpkg=./... -coverprofile coverage.out.tmp ./...
	(cat coverage.out.tmp | grep -v "mocks/" > coverage.out) && rm coverage.out.tmp
	cat coverage.out | gcov2lcov -outfile=coverage.lcov
	genhtml coverage.lcov -o $(COVERAGE_OUT_DIRECTORY) && open $(COVERAGE_OUT_DIRECTORY)/index.html

fmt:
	go fmt ./...