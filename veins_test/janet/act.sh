#!/bin/sh
set -e
chmod +x ./*.janet 2>/dev/null || true
./test.janet

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.janet

echo ""
echo "=== Test: Argument passing ==="
./echo_args.janet foo bar baz
