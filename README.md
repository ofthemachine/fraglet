# fraglet

**Run a fragment of code — in any of 100+ languages — inside an isolated container, from a shebang line, a CLI, or a Go import.**

A "fraglet" is an executable code file: a shebang plus a body. Point one at a container and the container runs *your* code instead of its own bundled "hello world" — no per-language SDK, no client library, no local interpreter install. Every language becomes a container with the same contract: mount code, run it, get stdout/stderr/exit-code (and optionally files) back.

## 30 seconds

```sh
curl -fsSL https://raw.githubusercontent.com/ofthemachine/fraglet/main/install.sh | sh
fragletc --vein=python -c 'print("hello from fraglet")'
```
```
hello from fraglet
```

Or as a standalone, directly-executable script:

```python
#!/usr/bin/env -S fragletc --vein=python
print("hello from fraglet")
```
```sh
chmod +x hello.py && ./hello.py
```

## Core concepts

| Term | Meaning |
|---|---|
| **fraglet** | An executable code file — shebang + body — that runs inside a container. |
| **vein** | A named language binding: `name → container image (+ file extensions)`. Built into the `fragletc` binary from [`pkg/embed/veins.yml`](pkg/embed/veins.yml), or bypassed entirely with `--image <any-registry-image>`. |
| **fraglet-entrypoint** | The binary baked into a fraglet-enabled container's `ENTRYPOINT`. Reads `fraglet.yaml`, injects or execs the mounted fraglet body. |
| **fragletc** | The host-side CLI. Resolves a vein or image, mounts your code, runs the container. |
| **mode** | A named variant of a container's execution config (different injection point, guide, or command) — e.g. `java:main` vs `java:wordalytica`. |
| **guide** / **essence** | Per-language (per-mode) docs shipped *inside* the container. `guide.md` is the full authoring reference; `essence.md` is a short, token-dense summary meant to be read first. |

## Three ways in

| Interface | Where | Consumer |
|---|---|---|
| CLI | `cmd/fragletc` | a human, a shell script, a shebang'd fraglet file |
| Go library | `pkg/runner`, `pkg/vein`, `pkg/engine`, `pkg/fraglet`, `pkg/dockercli` | other Go services importing this module directly, building their own orchestration on top instead of shelling out to `fragletc` |
| Container-side | `cmd/entrypoint` → `fraglet-entrypoint` | baked into a container's own `ENTRYPOINT`, making the image itself fraglet-enabled |

## Architecture

```
HOST                                             CONTAINER

fragletc (CLI)
  │
  ├─ resolves vein → image
  │    pkg/embed/veins.yml (built-in), or --image <any-registry-image>
  │
  └─ docker run -v code:/FRAGLET <image>  ─────▶  ENTRYPOINT = fraglet-entrypoint
                                                     reads fraglet.yaml
                                                     injects OR execs the mounted code
```

Two execution shapes, chosen per-container in `fraglet.yaml`:

- **Injection** (`execution.path` + `injection.match`) — the fraglet body replaces a marker line/region inside a real scaffold file (imports already wired up), and that file runs. What almost every [100hellos](#in-the-ecosystem) image uses.
- **Argv** (`execution.argv: true`) — for `FROM scratch` images with no shell or interpreter at all: a single static binary plus `fraglet-entrypoint`, nothing else in the image. The mounted body is treated as one command line, expanded to a literal argv, and exec'd directly against that binary.

## Install

Full instructions and troubleshooting: **[INSTALL.md](INSTALL.md)**.

```sh
curl -fsSL https://raw.githubusercontent.com/ofthemachine/fraglet/main/install.sh | sh
fragletc --vein=python -c 'print("hi")'
```

Docker is the only hard prerequisite — see INSTALL.md for platform-specific setup.

## Usage

```sh
# Run a file; extension infers the vein
fragletc script.py

# Any registry image, bypassing the built-in vein list entirely
fragletc --image ghcr.io/you/your-image:tag script.rb

# Parameters declared in the fraglet's own header (# fraglet-meta: param=...), passed as env vars
fragletc --param city=paris script.py

# Copy one declared output= file back to the host after a successful run
fragletc --output result.png render_meme.py

# Wrapping a tool whose output filename isn't known ahead of time: copy everything /output ends up with
fragletc --output-dir=./out wrap_a_real_cli.py

# What a language's fraglets need to know, shortest form first
fragletc essence python
fragletc guide python
```

`fragletc --help` documents every flag; `fragletc <subcommand> --help` (`refresh`, `guide`, `essence`) documents each subcommand.

## In the ecosystem

fraglet doesn't ship end-user containers itself — it's plumbing other projects build on, in two distinct ways:

- **Entrypoint-embedding.** A container bakes `fraglet-entrypoint` in as its `ENTRYPOINT` and ships a `fraglet.yaml`, becoming fraglet-enabled: runnable by name via `fragletc --vein=<name>`, or by `--image` for anything not in the built-in registry. **[100hellos](https://github.com/ofthemachine/100hellos)** is the reference example at scale — 100+ "hello world" language containers, 91 of them fraglet-enabled (each with a `fraglet/` directory: `fraglet.yaml` + `guide.md`). fraglet's [`pkg/embed/veins.yml`](pkg/embed/veins.yml) is the curated subset of those images `fragletc` knows about by name.
- **Library import.** A Go service imports `pkg/runner`, `pkg/vein`, `pkg/engine`, `pkg/fraglet`, and/or `pkg/dockercli` directly — rather than shelling out to the `fragletc` binary — to build its own orchestration on top: a ledger of runs, content-addressed memoization, its own CLI or agent-tool surface over `pkg/engine`. During active co-development, a consumer like this typically points its `go.mod` at a local fraglet checkout via a `replace` directive, tracking source directly rather than a pinned release.

## Releases

Three independent things ship from this repo, each on its own version line:

| What | Tag | Source of truth |
|---|---|---|
| The CLI binary | `fragletc-vX.Y.Z` | [`cmd/fragletc/releases/`](cmd/fragletc/releases/) |
| The binary containers embed | `entrypoint-vX.Y.Z` | [`cmd/entrypoint/releases/`](cmd/entrypoint/releases/) |
| The Go module | `vX.Y.Z` | `go.mod` — consumed by sibling repos, currently via `replace` directives rather than pinned versions during active co-development |

A push to `main` touching either `releases/` directory triggers that binary's build and GitHub Release automatically (`.github/workflows/release-*.yml`); nothing else does.

**Entrypoint changes ship to every fraglet-enabled container that exists.** There is no way to update one container without rebuilding all of them against the new binary. Treat `cmd/entrypoint`, `internal/entrypoint`, and `internal/executor` as the highest-blast-radius part of this repo — prefer a change scoped to `fragletc` (host-side only) whenever the goal doesn't strictly require touching the entrypoint.

## Repository layout

| Path | What |
|---|---|
| `cmd/fragletc/` | Host CLI entry point |
| `cmd/entrypoint/` | `fraglet-entrypoint` — the binary containers embed |
| `internal/entrypoint/`, `internal/executor/` | Entrypoint injection and execution internals |
| `pkg/vein/`, `pkg/embed/` | Vein model, registry, and the built-in embedded vein list |
| `pkg/engine/`, `pkg/runner/` | Container execution: `engine` orchestrates a run end-to-end, `runner` is the raw docker/local exec layer |
| `pkg/fraglet/` | Shared config schema; `fraglet-meta:` parsing (params, output decls); argv expansion for argv-mode |
| `pkg/guide/`, `pkg/essence/` | `guide`/`essence` subcommand implementations |
| `veins_test/` | Per-vein contract tests, by name, via the embedded registry |
| `cli_test/` | `fragletc` CLI behavior tests (flags, images, output handling) |

See [TESTING.md](TESTING.md) for the full test contract shared across this repo and 100hellos.

## Development

```sh
make build              # compile fragletc (embeds build-info.json)
make install            # build + copy fragletc onto $GOBIN
make test               # unit tests (go test -race ./...)
make test-veins         # veins_test/ against the embedded registry
make test-cli           # cli_test/ against fragletc's CLI surface
make test-entrypoint    # entrypoint injection/argv tests against real 100hellos images
make lint               # gofmt
make verify-100hellos LANGUAGE=<lang>  # run one 100hellos image's own fraglet/verify.sh
```

Session and agent guidance lives in [CLAUDE.md](CLAUDE.md).
