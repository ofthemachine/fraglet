#!/bin/sh
set -e
chmod +x ./*.zig 2>/dev/null || true
./test.zig

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.zig

echo ""
echo "=== Test: Argument passing ==="
./echo_args.zig foo bar baz
