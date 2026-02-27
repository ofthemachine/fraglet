#!/bin/sh
set -e
chmod +x ./*.apl 2>/dev/null || true
./test.apl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.apl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.apl foo bar baz
