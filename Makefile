.PHONY: run test lint build build-cli install install-cli test-integration test-harness test-entrypoint test-envelopes

# Copy envelopes for embedding
pkg/embed/envelopes:
	@mkdir -p pkg/embed/envelopes
	@cp envelopes/*.yml pkg/embed/envelopes/

build: pkg/embed/envelopes
	go build -o fraglet-mcp .

build-cli: pkg/embed/envelopes
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

test-envelopes:
	cd envelopes_test && go test -tags=integration -v .

test-cli:
	cd cli_test && go test -tags=integration -v .

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
