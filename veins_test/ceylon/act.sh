#!/bin/sh
set -e
chmod +x ./*.ceylon 2>/dev/null || true
./test.ceylon

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ceylon

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ceylon foo bar baz
