# Envelopes Test Suite

This directory contains integration tests for fraglet envelopes by name using `fragletc`.

## Purpose

These tests verify that each envelope works correctly when referenced by name. All tests use the `--envelope` flag to reference envelopes from the embedded registry, not direct container images.

## Structure

Each language has its own test directory:
```
envelopes_test/
  <language>/
    act.sh      - Test script that runs fragletc commands
    assert.txt  - Expected output for assertions
  envelopes_test.go - Test harness using clitest
```

## Test Format

### act.sh
- Executable shell script that runs `fragletc` commands
- The `fragletc` binary is provided by the clitest harness with embedded envelopes
- **Must use `--envelope <name>` to reference envelopes by name** (not `--image`)
- Keep it simple: test basic execution, multi-line code, and file input
- Should be idempotent and clean up any temporary files
- No need to test fraglet-path explicitly (that's part of envelope config)

### assert.txt
- Contains the expected output from running `act.sh`
- Used for automated assertion checking by the clitest harness
- Should match the actual output exactly

## Running Tests

### Automated Testing (Recommended)
```bash
make test-envelopes
```

Or directly:
```bash
cd envelopes_test
go test -tags=integration -v .
```

### Manual Testing
```bash
cd envelopes_test/<language>
# First ensure fragletc is built and copied
cd .. && make build-cli && cp ../fragletc . && cd <language>
./act.sh
```

## Adding New Language Tests

1. Create a new directory: `envelopes_test/<language>/`
2. Create `act.sh` with fragletc test commands (use `./fragletc` - it's provided by the harness)
3. Run the test: `make test-envelopes`
4. The harness will capture output and compare with `assert.txt`
5. If output differs, update `assert.txt` with the correct expected output

## Example

See `python/` for a complete example with:
- Basic execution (STDIN)
- Multi-line code
- File input

The test is intentionally minimal - it just verifies the envelope works correctly by name.

## Notes

- Envelopes are embedded in the built `fragletc` binary (from `envelopes/` directory)
- The test harness automatically builds `fragletc` with embedded envelopes before running tests
- Tests run in parallel by default (controlled by the clitest harness)
- **All tests use envelope names, not container images** - this verifies the envelope registry works correctly

## Difference from cli_test

- **envelopes_test**: Tests envelope correctness by name (uses `--envelope` flag)
- **cli_test**: Tests the CLI tool itself with direct container images and embedded envelopes

