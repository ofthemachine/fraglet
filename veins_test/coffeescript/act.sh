#!/bin/sh
set -e
chmod +x ./*.coffee 2>/dev/null || true
./test.coffee

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.coffee

echo ""
echo "=== Test: Argument passing ==="
./echo_args.coffee foo bar baz
