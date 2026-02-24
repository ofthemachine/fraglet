#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.vb 2>/dev/null || true
./test.vb

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.vb

echo ""
echo "=== Test: Argument passing ==="
./echo_args.vb foo bar baz
./echo_args.vb "queen foo" 1 2

echo ""
echo "=== Test: Visual Basic fortune ==="
./visual_basic_fortune.vb
