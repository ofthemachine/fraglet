#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.apl 2>/dev/null || true
./test.apl

echo ""
echo "=== Fibonacci (many-line) ==="
./fib.apl

echo ""
echo "=== FizzBuzz (many-line) ==="
./fizzbuzz.apl

echo ""
echo "=== 5×5 times table (many-line) ==="
./times_table.apl
