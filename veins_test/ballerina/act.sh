#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.bal 2>/dev/null || true
./test.bal

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.bal

echo ""
echo "=== Test: Argument passing ==="
./echo_args.bal foo bar baz
