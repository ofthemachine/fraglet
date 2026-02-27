#!/bin/sh
set -e
chmod +x ./*.rs 2>/dev/null || true
./test.rs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.rs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.rs foo bar baz
