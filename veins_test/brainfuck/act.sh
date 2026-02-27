#!/bin/sh
set -e
chmod +x ./*.bf 2>/dev/null || true
./test.bf

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.bf
