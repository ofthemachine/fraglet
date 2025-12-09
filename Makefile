.PHONY: run test lint build build-cli install install-cli test-integration test-harness test-entrypoint test-envelopes

# Copy envelopes for embedding
# Use a stamp file to track when copy was last done
pkg/embed/envelopes/.stamp: $(wildcard envelopes/*.yml)
	@mkdir -p pkg/embed/envelopes
	@cp envelopes/*.yml pkg/embed/envelopes/
	@touch pkg/embed/envelopes/.stamp

# Phony target that depends on stamp file
pkg/embed/envelopes: pkg/embed/envelopes/.stamp

build: pkg/embed/envelopes
	go build -o fraglet-mcp .

build-cli: pkg/embed/envelopes
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

test-envelopes:
	cd envelopes_test && go test -tags=integration -v .

test-cli:
	cd cli_test && go test -tags=integration -v .

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
