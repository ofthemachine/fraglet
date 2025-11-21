.PHONY: run test lint build install test-integration test-harness test-entrypoint

build:
	go build -o fraglet-mcp .

install:
	go install fraglet-mcp

test:
	go test ./...

test-entrypoint:
	cd entrypoint && go test -tags=integration -v ./tests/...

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
