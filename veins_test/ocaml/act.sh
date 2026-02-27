#!/bin/sh
set -e
chmod +x ./*.ml 2>/dev/null || true
./test.ml

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ml

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ml foo bar baz
