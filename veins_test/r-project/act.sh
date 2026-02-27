#!/bin/sh
set -e
chmod +x ./*.r 2>/dev/null || true
./test.r

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.r

echo ""
echo "=== Test: Argument passing ==="
./echo_args.r foo bar baz
