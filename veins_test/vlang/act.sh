#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.v 2>/dev/null || true
./test.v

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.v

echo ""
echo "=== Test: Argument passing ==="
./echo_args.v foo bar baz

echo ""
echo "=== Test: V fortune ==="
./vlang_fortune.v
