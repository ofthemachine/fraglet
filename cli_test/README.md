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
- Fraglet path configuration (`--fraglet-path` / `-p`)
- Error handling and validation

## Structure

Each test category has its own directory:
```
cli_test/
  stdin/      - STDIN input tests
  file/       - File input tests
  vein/       - Embedded vein tests
  errors/     - Error handling tests
  cli_test.go - Test harness using clitest
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


