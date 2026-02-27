#!/bin/sh
set -e
chmod +x ./*.chpl 2>/dev/null || true
./test.chpl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.chpl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.chpl foo bar baz
