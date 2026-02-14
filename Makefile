.PHONY: run test lint build build-cli install install-cli test-integration test-harness test-entrypoint test-veins verify-100hellos

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

# Run 100hellos fraglet verify.sh for a language. Requires HELLOS_ROOT (default: $HOME/repos/100hellos).
# Usage: make verify-100hellos LANGUAGE=ats
HELLOS_ROOT ?= $(HOME)/repos/100hellos
verify-100hellos:
	@if [ -z "$(LANGUAGE)" ]; then echo "Usage: make verify-100hellos LANGUAGE=<lang>"; exit 1; fi; \
	if [ ! -d "$(HELLOS_ROOT)/$(LANGUAGE)/fraglet" ]; then echo "Error: $(HELLOS_ROOT)/$(LANGUAGE)/fraglet not found"; exit 1; fi; \
	$(MAKE) install-cli && \
	"$(HELLOS_ROOT)/$(LANGUAGE)/fraglet/verify.sh" "100hellos/$(LANGUAGE):local"

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
