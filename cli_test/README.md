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

## Structure

Each test category has its own directory:
```
cli_test/
  stdin/         - STDIN input tests
  file/          - File input tests
  vein/          - Embedded vein tests
  fraglet_help/  - --fraglet-help + fraglet-meta (multi-scenario act/assert)
  errors/        - Error handling tests
  cli_test.go    - Test harness using clitest
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


