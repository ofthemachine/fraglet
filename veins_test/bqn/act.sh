#!/bin/sh
set -e
chmod +x ./*.bqn 2>/dev/null || true
./test.bqn

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.bqn

echo ""
echo "=== Test: Argument passing ==="
./echo_args.bqn foo bar baz
