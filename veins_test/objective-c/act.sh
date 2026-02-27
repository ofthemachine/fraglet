#!/bin/sh
set -e
chmod +x ./*.m 2>/dev/null || true
./test.m

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.m

echo ""
echo "=== Test: Argument passing ==="
./echo_args.m foo bar baz
