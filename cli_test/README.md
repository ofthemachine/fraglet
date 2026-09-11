# CLI Test Suite

This directory contains integration tests for the `fragletc` CLI tool itself, testing direct container/image usage and CLI functionality.

## Purpose

These tests verify:
- CLI flag parsing and validation
- Direct container image execution (`--image` / `-i`)
- Embedded vein execution (`--vein` / `-v`)
- File input handling (positional arguments)
- Extension-to-vein inference
- Vein mode syntax (`vein:mode`)
- Fraglet path configuration (`--fraglet-path`; long form only)
- `--fraglet-help` and `fraglet-meta:` parameter declarations (shebang files + `-c`, dedup, errors); `-p` / `--param` / `--fraglet-help` stripped from argv anywhere before `--`
- Error handling and validation
- `output=` declarations + `--output <relpath>[=hostdest]`: `/output` mounts whenever any output= is declared (independent of `--output`); every declared output not named by `--output` is reported as discarded, never silently dropped or silently kept
- `--output-dir <hostdir>`: copies everything written to `/output` into `hostdir` after the run, with no `output=` declaration needed — for a fraglet whose output filename is only known at runtime; mutually exclusive with `--output`
- `--output <dest>` shorthand: a bare, unrecognized relpath with exactly one `output=` declared is treated as that output's destination — no need to repeat the fraglet's own declared filename when there's nothing to disambiguate; left alone (real "not declared" error) when >1 output is declared or the caller wrote an explicit `relpath=dest`
- `param=<alias>:file`: a file-shaped param's CLI value is a host path, mounted read-only at the fixed container path `/input/<alias>` (never the host path itself); multiple file params mount independently; a missing host file fails before the container starts
- `param=<alias>:required` is enforced host-side before any container starts: a missing required param prints the same listing `--fraglet-help` shows and exits 2, rather than silently expanding to `""` wherever the fraglet body references it and letting the wrapped tool's own (often confusing) error surface instead. `default=` on the same decl exempts it — a default already satisfies "the caller must supply this" (`fraglet.MissingRequired`, shared with operon's own `internal/runparams`, so the exemption rule can't drift between the two). `required` is checked only against `-p`/`--param` values; a value supplied via raw `-e` env-forwarding is invisible to it by design (`-e` and `param=` are deliberately separate mechanisms).
- `default=` is injected into the container env when the caller omits `-p` for that alias (`fraglet.ApplyDefaults`). Explicit `-p alias=` (including empty) always wins and is never overwritten. `description=` / short `d=` on a param token (must be last) surfaces in `--fraglet-help` and the missing-required listing; values may contain spaces. See `cli_test/param_defaults/`.
- Relative output paths (`--output-dir=.` baked into a shebang, `--output=<dest>`) resolve against the caller's cwd at invocation, never the directory the script itself lives in — a self-exec script at `skills/fun/meme.sh` behaves like a locally installed tool, not like its output depends on where it's checked into the repo
- Container network isolation: `--network none` and `#: network=none` launch Docker with `--network none` (no `eth0`); `#: network=required` and an explicit `--network` flag override that default. Docker's default bridge is used when neither the flag nor the header asks for isolation.

## Structure

Each test category has its own directory:
```
cli_test/
  stdin/            - STDIN input tests
  file/             - File input tests
  vein/             - Embedded vein tests
  fraglet_help/     - --fraglet-help + fraglet-meta (multi-scenario act/assert)
  errors/           - Error handling tests
  output_declared/  - output= + --output (no-flag POLA case, partial requests, full requests)
  output_dir/       - --output-dir (runtime-chosen filenames, default naming, mutual exclusion with --output)
  output_shorthand/ - --output <dest> single-output shorthand (no relpath needed when unambiguous)
  output_relative_to_caller/ - relative --output-dir/--output resolve against the caller's cwd, not the script's own directory
  param_file/       - param=<alias>:file (host file mounted read-only at /input/<alias>)
  required_params/  - param=<alias>:required enforced before any container starts
  param_defaults/   - default= injection + param description= in --fraglet-help
  network/          - container network isolation (--network, #: network=none/required)
  cli_test.go       - Test harness using clitest
```

## Test Format

### act.sh
- Executable shell script that runs `fragletc` commands
- The `fragletc` binary is provided by the clitest harness
- Tests CLI functionality, not vein correctness

### assert.txt
- Contains the expected output from running `act.sh`
- Used for automated assertion checking

## Running Tests

```bash
make test-cli
```

Or directly:
```bash
cd cli_test
go test -tags=integration -v .
```

## Adding New Tests

1. Create a new directory: `cli_test/<category>/`
2. Create `act.sh` with fragletc test commands
3. Run the test: `make test-cli`
4. Update `assert.txt` with correct expected output

## Difference from veins_test

- **cli_test**: Tests the CLI tool itself with direct container images and embedded veins
- **veins_test**: Tests vein correctness by name (uses embedded veins only)

## Difference from entrypoint tests

- **`--param` / `-p`** plus **FRAGLET_PARAM_* → bare env** (coerce + strip) is exercised in [`entrypoint/tests/params_coerce`](../entrypoint/tests/params_coerce): that suite **`docker build`s a small test image** whose Dockerfile **`COPY`s a locally compiled `fraglet-entrypoint` binary** and sets it as `ENTRYPOINT`. **`cli_test` `inline_code`** uses raw `--image` (typical `docker run … sh -c`); it does not layer or invoke that binary, so it does not assert param transport semantics.


