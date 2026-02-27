#!/bin/sh
set -e
chmod +x ./*.asm 2>/dev/null || true
./test.asm

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.asm

echo ""
echo "=== Test: Argument passing ==="
./echo_args.asm foo bar baz
