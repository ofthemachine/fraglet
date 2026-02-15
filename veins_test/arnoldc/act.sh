#!/bin/sh
# Use local image so stdin works
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.arnoldc 2>/dev/null || true

./test.arnoldc

echo ""
echo "=== Test: Quotes (multi-line) ==="
./quotes.arnoldc

echo ""
echo "=== Test: Math (5 + 10) ==="
./math.arnoldc

echo ""
echo "=== Test: Stdin (integer read, doubled) ==="
echo "42" | ./stdin_int.arnoldc
