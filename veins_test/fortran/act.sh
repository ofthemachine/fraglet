#!/bin/sh
set -e
chmod +x ./*.f 2>/dev/null || true
./test.f

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.f

echo ""
echo "=== Test: Argument passing ==="
./echo_args.f foo bar baz
