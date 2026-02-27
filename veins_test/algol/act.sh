#!/bin/sh
set -e
chmod +x ./*.alg 2>/dev/null || true
./test.alg

echo ""
echo "=== Test: Stdin (integer read) ==="
echo "42" | ./stdin_int.alg
