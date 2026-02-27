#!/bin/sh
set -e
chmod +x ./*.sed 2>/dev/null || true
./test.sed

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.sed
