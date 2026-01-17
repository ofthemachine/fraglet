.PHONY: run test lint build build-cli install install-cli test-integration test-harness test-entrypoint test-veins

# Veins are embedded directly from pkg/embed/veins.yml via go:embed
# No build-time copying needed

build:
	go build -o fraglet-mcp .

build-cli:
	go build -o fragletc ./cli

install: build
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; \
	cp fraglet-mcp $$GOBIN/

install-cli: build-cli
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; \
	cp fragletc $$GOBIN/

test:
	go test ./...

test-entrypoint:
	cd entrypoint && go test -tags=integration -v ./tests/...

test-veins:
	cd veins_test && go test -tags=integration -v .

test-cli:
	cd cli_test && go test -tags=integration -v .

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
