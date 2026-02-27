#!/bin/sh
set -e
chmod +x ./*.d 2>/dev/null || true
./test.d

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.d

echo ""
echo "=== Test: Argument passing ==="
./echo_args.d foo bar baz
