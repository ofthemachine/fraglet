#!/bin/sh
set -e
chmod +x ./*.odin 2>/dev/null || true
./test.odin

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.odin

echo ""
echo "=== Test: Argument passing ==="
./echo_args.odin foo bar baz
