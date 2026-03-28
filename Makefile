.PHONY: run test lint build install build-info test-integration test-harness test-entrypoint test-veins verify-100hellos

# Veins are embedded directly from pkg/embed/veins.yml via go:embed
# No build-time copying needed

# Reproducible build flags — same source + same Go version = identical binary.
# -trimpath: strips host-specific paths from the binary
# -buildvcs=false: we embed our own provenance via build-info.json
# -s -w: strip symbol table and DWARF (smaller binary, matches CI)
GO_BUILD_FLAGS = -trimpath -buildvcs=false -ldflags="-s -w"

CLITEST_PROGRESS ?= $(if $(CI),0,1)

build-info:
	@VERSION=$$(ls cmd/fragletc/releases/*.md 2>/dev/null | xargs -I{} basename {} .md | sort -V | tail -1); \
	[ -z "$$VERSION" ] && VERSION="dev"; \
	COMMIT=$$(git rev-parse HEAD 2>/dev/null || echo "unknown"); \
	BUILD_TIME=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	DIRTY=$$([ -z "$$(git status --porcelain 2>/dev/null)" ] && echo "false" || echo "true"); \
	printf '{\n  "version": "%s",\n  "commit": "%s",\n  "buildTime": "%s",\n  "dirty": %s\n}\n' \
		"$$VERSION" "$$COMMIT" "$$BUILD_TIME" "$$DIRTY" > cmd/fragletc/build-info.json

build: build-info
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o fragletc ./cmd/fragletc

build-entrypoint:
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o fraglet-entrypoint ./cmd/entrypoint

install: build
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; \
	cp fragletc $$GOBIN/

test:
	go test ./...

test-entrypoint:
	CLITEST_PROGRESS=$(CLITEST_PROGRESS) cd cmd/entrypoint && go test -tags=integration -v ./tests/...

# Slow/heavy target; prefer test-entrypoint or test-cli for quick clitest feedback.
test-veins:
	cd veins_test && go test -tags=integration -v .

test-cli: build-info
	CLITEST_PROGRESS=$(CLITEST_PROGRESS) cd cli_test && go test -tags=integration -v .

# Run 100hellos fraglet verify.sh for a language. Requires HELLOS_ROOT (default: $HOME/repos/100hellos).
# Usage: make verify-100hellos LANGUAGE=ats
HELLOS_ROOT ?= $(HOME)/repos/100hellos
verify-100hellos:
	@if [ -z "$(LANGUAGE)" ]; then echo "Usage: make verify-100hellos LANGUAGE=<lang>"; exit 1; fi; \
	if [ ! -d "$(HELLOS_ROOT)/$(LANGUAGE)/fraglet" ]; then echo "Error: $(HELLOS_ROOT)/$(LANGUAGE)/fraglet not found"; exit 1; fi; \
	$(MAKE) install && \
	"$(HELLOS_ROOT)/$(LANGUAGE)/fraglet/verify.sh" "100hellos/$(LANGUAGE):local"

lint:
	find . -name "*.go" | grep -v './vendor' | xargs -I{} go fmt {}
