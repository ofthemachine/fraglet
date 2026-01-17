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
  generate.sh    - Generator script to sync from 100hellos
```

## Test Format

### Code Scripts (Shebang Pattern)

Test code is written as executable scripts with shebang lines:

```python
#!/usr/bin/env -S fragletc --vein=python
print("Hello, World!")
```

**Shebang Format**: `#!/usr/bin/env -S fragletc --vein=<vein-name>`

This format enables:
- **In test harness**: fragletc is in temp dir, temp dir is on PATH, so scripts execute directly
- **Direct execution**: Uses system-installed fragletc from PATH when run outside the harness

The `-S` flag to `env` allows splitting arguments (requires GNU coreutils 8.30+).

### act.sh

Minimal shell script that runs the test scripts:

```sh
#!/bin/sh
set -e
chmod +x ./*.py 2>/dev/null || true
./hello.py
```

For multiple tests per vein:
```sh
#!/bin/sh
set -e
chmod +x ./*.py 2>/dev/null || true

echo "=== Test 1: Basic execution ==="
./hello.py

echo ""
echo "=== Test 2: Multi-line code ==="
./multiline.py
```

### assert.txt

Contains the expected output from running `act.sh`. Used for automated assertion checking.

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

You can run tests directly (if fragletc is installed):
```bash
cd veins_test/python
./hello.py
```

Or via the harness:
```bash
cd veins_test/python
./act.sh
```

## Adding New Tests

### Manual Creation

1. Create a new directory: `veins_test/<language>/`
2. Create code script(s) with shebang: `#!/usr/bin/env -S fragletc --vein=<name>`
3. Create `act.sh` that runs the script(s)
4. Run the test: `make test-veins`
5. Update `assert.txt` with correct expected output

### Using Generator

The `generate.sh` script can generate tests from 100hellos sources:

```bash
./generate.sh elixir          # Generate test for elixir
./generate.sh --all           # Generate all available
./generate.sh --sync          # Update existing from 100hellos
```

**Source priority**:
1. `100hellos/{lang}/fraglet/verify.sh` - Extract first example
2. `100hellos/{lang}/fraglet/guide.md` - Parse first code block from Examples section
3. `100hellos/{lang}/files/hello-world.*` - Strip shebang, use as fallback

The generator:
- Determines file extension from `pkg/embed/veins.yml`
- Creates shebang scripts with code
- Generates minimal `act.sh`
- Attempts to generate `assert.txt` by running the script (requires fragletc)

## Multiple Tests Per Vein

Some veins have multiple test scripts for different scenarios:

```
python/
  hello.py           # Required: simple smoke test
  multiline.py       # Optional: additional coverage
  file_input.py      # Optional: file handling
  act.sh             # Runs all scripts
  assert.txt
```

## Argument Passing Tests

To test argument passing through the shebang → fragletc → container pipeline:

**`python/echo_args.py`**:
```python
#!/usr/bin/env -S fragletc --vein=python
import sys
print(f"Args: {sys.argv[1:]}")
```

**In `python/act.sh`**:
```sh
./echo_args.py foo bar baz
```

This verifies the full argument pipeline works correctly.

## Notes

- Veins are embedded in the built `fragletc` binary (from `pkg/embed/veins.yml`)
- The test harness automatically builds `fragletc` with embedded veins before running tests
- **All tests use vein names, not container images** - this verifies the vein registry works correctly
- Tests can use extension inference (e.g., `fragletc script.py` instead of `fragletc --vein python script.py`) for special cases
- The shebang pattern (`#!/usr/bin/env -S fragletc --vein=X`) is the most inflexible interface and defines the parameter-space of fragletc

## Difference from cli_test

- **veins_test**: Tests vein correctness by name (uses `--vein` flag via shebang)
- **cli_test**: Tests the CLI tool itself with direct container images and embedded veins

## Generator Script

The `generate.sh` script helps maintain consistency and enables syncing from 100hellos:

- **Maintainability**: Single source of truth for test generation
- **Scalability**: Can generate tests for 100+ languages
- **Evolution**: Easy to update tests when 100hellos examples change
- **Separation**: Tests remain contract tests, independent of 100hellos implementation details

Set `HELLOS_ROOT` environment variable to override the default `$HOME/repos/100hellos` path.
