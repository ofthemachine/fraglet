#!/bin/sh
set -e
chmod +x ./*.fs 2>/dev/null || true
./test.fs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.fs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.fs foo bar baz
