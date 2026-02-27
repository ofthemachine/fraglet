#!/bin/sh
set -e
chmod +x ./*.factor 2>/dev/null || true
./test.factor

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.factor

echo ""
echo "=== Test: Argument passing ==="
./echo_args.factor foo bar baz
