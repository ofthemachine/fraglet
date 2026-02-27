#!/bin/sh
set -e
chmod +x ./*.io 2>/dev/null || true
./test.io

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.io

echo ""
echo "=== Test: Argument passing ==="
./echo_args.io foo bar baz
