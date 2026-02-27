#!/bin/sh
set -e
chmod +x ./*.dart 2>/dev/null || true
./test.dart

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.dart

echo ""
echo "=== Test: Argument passing ==="
./echo_args.dart foo bar baz
