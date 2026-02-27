#!/bin/sh
set -e
chmod +x ./*.vb 2>/dev/null || true
./test.vb

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.vb

echo ""
echo "=== Test: Argument passing ==="
./echo_args.vb foo bar baz
