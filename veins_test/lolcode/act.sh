#!/bin/sh
set -e
chmod +x ./*.lol 2>/dev/null || true
./test.lol

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.lol
