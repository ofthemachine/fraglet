# CLI Test Suite

This directory contains integration tests for the `fragletc` CLI tool itself, testing direct container/image usage and CLI functionality.

## Purpose

These tests verify:
- CLI flag parsing and validation
- Direct container image execution (`--image` / `-i`)
- Embedded envelope execution (`--envelope` / `-e`)
- File input handling (`--input` / `-f`)
- Fraglet path configuration (`--fraglet-path` / `-p`)
- Error handling and validation

## Structure

Each test category has its own directory:
```
cli_test/
  stdin/      - STDIN input tests
  file/       - File input tests
  envelope/   - Embedded envelope tests
  errors/     - Error handling tests
  cli_test.go - Test harness using clitest
```

## Test Format

### act.sh
- Executable shell script that runs `fragletc` commands
- The `fragletc` binary is provided by the clitest harness
- Tests CLI functionality, not envelope correctness

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

## Difference from envelopes_test

- **cli_test**: Tests the CLI tool itself with direct container images and embedded envelopes
- **envelopes_test**: Tests envelope correctness by name (uses embedded envelopes only)


