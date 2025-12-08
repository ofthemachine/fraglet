.PHONY: run test lint build build-cli install install-cli test-integration test-harness test-entrypoint

build:
	go build -o fraglet-mcp .

build-cli:
	go build -o fragletc ./cli

install:
	go install fraglet-mcp

install-cli:
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; \
	go build -o $$GOBIN/fragletc ./cli

test:
	go test ./...

test-entrypoint:
	cd entrypoint && go test -tags=integration -v ./tests/...

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
