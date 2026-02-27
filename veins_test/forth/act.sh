#!/bin/sh
set -e
chmod +x ./*.fth 2>/dev/null || true
./test.fth

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.fth

echo ""
echo "=== Test: Argument passing ==="
./echo_args.fth foo bar baz
