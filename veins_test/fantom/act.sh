#!/bin/sh
set -e
chmod +x ./*.fan 2>/dev/null || true
./test.fan

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.fan

echo ""
echo "=== Test: Argument passing ==="
./echo_args.fan foo bar baz
