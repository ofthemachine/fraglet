#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.hs 2>/dev/null || true
./test.hs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.hs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.hs foo bar baz
./echo_args.hs "queen foo" 1 2

echo ""
echo "=== Test: Haskell fortune ==="
./haskell_fortune.hs
