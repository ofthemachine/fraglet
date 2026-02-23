#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.ceylon 2>/dev/null || true
./test.ceylon

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ceylon

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ceylon foo bar baz
./echo_args.ceylon "queen foo" 1 2

echo ""
echo "=== Test: Ceylon fortune ==="
./ceylon_fortune.ceylon
