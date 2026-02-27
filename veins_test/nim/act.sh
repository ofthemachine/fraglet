#!/bin/sh
set -e
chmod +x ./*.nim 2>/dev/null || true
./test.nim

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.nim

echo ""
echo "=== Test: Argument passing ==="
./echo_args.nim foo bar baz
