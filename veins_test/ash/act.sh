#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.ash 2>/dev/null || true
./test.ash

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ash

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ash foo bar baz
