#!/bin/sh
set -e
chmod +x ./*.idr 2>/dev/null || true
./test.idr

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.idr

echo ""
echo "=== Test: Argument passing ==="
./echo_args.idr foo bar baz
