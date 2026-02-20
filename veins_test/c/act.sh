#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.c 2>/dev/null || true
./test.c

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.c

echo ""
echo "=== Test: Argument passing ==="
./echo_args.c foo bar baz

echo ""
echo "=== Test: Array sum ==="
./array_sum.c

echo ""
echo "=== Test: C fortune ==="
./c_fortune.c
