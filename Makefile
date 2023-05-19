.PHONY: all
all: docs fmt vet test

.PHONY: test
test: fmt vet unit-test

.PHONY: fmt
fmt:
	go fmt -C test ./...

.PHONY: vet
vet:
	go vet -C test ./...

.PHONY: unit-test
unit-test:
	go test -C test ./... -count=1
