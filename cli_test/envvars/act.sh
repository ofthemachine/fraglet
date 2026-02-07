#!/bin/sh
set -e

# Test fragletc environment variable passthrough

echo "=== Test 1: Forward host env var ==="
export FRAGLET_TEST_HOST_VAR="hello_from_host"
fragletc --vein python -e FRAGLET_TEST_HOST_VAR -c 'import os; print(os.environ.get("FRAGLET_TEST_HOST_VAR", "NOT_SET"))'

echo ""
echo "=== Test 2: Explicit KEY=VALUE ==="
fragletc --vein python -e MY_VAR=explicit_value -c 'import os; print(os.environ.get("MY_VAR", "NOT_SET"))'

echo ""
echo "=== Test 3: Missing host var (silently skipped) ==="
fragletc --vein python -e DEFINITELY_NOT_SET -c 'import os; print(os.environ.get("DEFINITELY_NOT_SET", "correctly_absent"))'

echo ""
echo "=== Test 4: Multiple -e flags ==="
export FRAGLET_A="alpha"
export FRAGLET_B="beta"
fragletc --vein python -e FRAGLET_A -e FRAGLET_B -e FRAGLET_C=gamma -c 'import os; print(os.environ.get("FRAGLET_A")); print(os.environ.get("FRAGLET_B")); print(os.environ.get("FRAGLET_C"))'

echo ""
echo "=== Test 5: Env vars don't leak without -e ==="
export SECRET_KEY="should_not_appear"
fragletc --vein python -c 'import os; print(os.environ.get("SECRET_KEY", "correctly_absent"))'
