#!/bin/sh
set -e
chmod +x ./*.cs 2>/dev/null || true
./test.cs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.cs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.cs foo bar baz
