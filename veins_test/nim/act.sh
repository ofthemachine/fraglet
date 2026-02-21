#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.nim 2>/dev/null || true
./test.nim

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.nim

echo ""
echo "=== Test: Argument passing ==="
./echo_args.nim foo bar baz
./echo_args.nim "queen foo" 1 2

echo ""
echo "=== Test: Nim fortune ==="
./nim_fortune.nim
