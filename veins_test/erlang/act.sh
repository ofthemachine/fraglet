#!/bin/sh
set -e
chmod +x ./*.erl 2>/dev/null || true
./test.erl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.erl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.erl foo bar baz
