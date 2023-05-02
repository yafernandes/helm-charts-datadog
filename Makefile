.PHONY: all
all: docs fmt vet test

.PHONY: docs
docs:
	./.github/helm-docs.sh

.PHONY: test
test: fmt vet chart-test

.PHONY: fmt
fmt:
	go fmt -C test ./...

.PHONY: vet
vet:
	go vet -C test ./...

.PHONY: chart-test
chart-test:
	go test -C test ./... -count=1

.PHONY: integration-tests
integration-tests:
	go test -C test/integ --tags=integration -count=1 -v

