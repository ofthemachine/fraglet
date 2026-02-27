#!/bin/sh
set -e
chmod +x ./*.alg 2>/dev/null || true
./test.alg

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.alg
