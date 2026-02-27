#!/bin/sh
set -e
chmod +x ./*.c 2>/dev/null || true
./test.c

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.c

echo ""
echo "=== Test: Argument passing ==="
./echo_args.c foo bar baz
