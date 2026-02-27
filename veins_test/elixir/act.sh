#!/bin/sh
set -e
chmod +x ./*.exs 2>/dev/null || true
./test.exs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.exs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.exs foo bar baz
