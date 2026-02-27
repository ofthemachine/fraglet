#!/bin/sh
set -e
chmod +x ./*.lua 2>/dev/null || true
./test.lua

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.lua

echo ""
echo "=== Test: Argument passing ==="
./echo_args.lua foo bar baz
