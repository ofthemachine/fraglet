#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.fan 2>/dev/null || true
./test.fan

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.fan

echo ""
echo "=== Test: Argument passing ==="
./echo_args.fan foo bar baz
./echo_args.fan "queen foo" 1 2

echo ""
echo "=== Test: Fantom fortune ==="
./fantom_fortune.fan
