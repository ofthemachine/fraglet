# Veins Test Suite

This directory contains integration tests for fraglet veins by name using `fragletc`.

## Purpose

These tests verify that each vein works correctly when referenced by name. All tests use the `--vein` flag to reference veins from the embedded registry, not direct container images.

## Structure

Each language has its own directory:
```
veins_test/
  python/     - Python vein tests
  javascript/ - JavaScript vein tests
  c/          - C programming language vein tests
  ...
  veins_test.go - Test harness using clitest
```

## Test Format

### act.sh
- Executable shell script that runs `fragletc` commands
- The `fragletc` binary is provided by the clitest harness with embedded veins
- **Must use `--vein <name>` to reference veins by name** (not `--image`)
- Tests vein correctness, not CLI functionality

### assert.txt
- Contains the expected output from running `act.sh`
- Used for automated assertion checking

## Running Tests

```bash
make test-veins
```

Or directly:
```bash
cd veins_test
go test -tags=integration -v .
```

## Running Individual Tests

```bash
cd veins_test/<language>
./act.sh
```

## Adding New Tests

1. Create a new directory: `veins_test/<language>/`
2. Create `act.sh` with fragletc test commands using `--vein <name>`
3. Run the test: `make test-veins`
4. Update `assert.txt` with correct expected output

The test is intentionally minimal - it just verifies the vein works correctly by name.

## Notes

- Veins are embedded in the built `fragletc` binary (from `pkg/embed/veins.yml`)
- The test harness automatically builds `fragletc` with embedded veins before running tests
- **All tests use vein names, not container images** - this verifies the vein registry works correctly
- Tests can use extension inference (e.g., `fragletc script.py` instead of `fragletc --vein python script.py`)

## Difference from cli_test

- **veins_test**: Tests vein correctness by name (uses `--vein` flag)
- **cli_test**: Tests the CLI tool itself with direct container images and embedded veins
