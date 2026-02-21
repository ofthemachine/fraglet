#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.mksh 2>/dev/null || true
./test.mksh

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.mksh

echo ""
echo "=== Test: Argument passing ==="
./echo_args.mksh foo bar baz
./echo_args.mksh "queen foo" 1 2

echo ""
echo "=== Test: mksh fortune ==="
./mksh_fortune.mksh
