#!/bin/sh
set -e
chmod +x ./*.vala 2>/dev/null || true
./test.vala

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.vala

echo ""
echo "=== Test: Argument passing ==="
./echo_args.vala foo bar baz
