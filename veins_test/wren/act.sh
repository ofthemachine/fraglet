#!/bin/sh
set -e
chmod +x ./*.wren 2>/dev/null || true
./test.wren

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.wren

echo ""
echo "=== Test: Argument passing ==="
./echo_args.wren foo bar baz
