#!/bin/sh
set -e
chmod +x ./*.adb 2>/dev/null || true
./test.adb

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.adb

echo ""
echo "=== Test: Argument passing ==="
./echo_args.adb foo bar baz
