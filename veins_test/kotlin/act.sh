#!/bin/sh
set -e
chmod +x ./*.kt 2>/dev/null || true
./test.kt

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.kt

echo ""
echo "=== Test: Argument passing ==="
./echo_args.kt foo bar baz
