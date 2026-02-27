#!/bin/sh
set -e
chmod +x ./*.l 2>/dev/null || true
./test.l

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.l

echo ""
echo "=== Test: Argument passing ==="
./echo_args.l foo bar baz
