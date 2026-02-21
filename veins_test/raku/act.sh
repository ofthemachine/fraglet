#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.raku 2>/dev/null || true
./test.raku

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.raku

echo ""
echo "=== Test: Argument passing ==="
./echo_args.raku foo bar baz
./echo_args.raku "queen foo" 1 2

echo ""
echo "=== Test: Raku fortune ==="
./raku_fortune.raku
