#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.bqn 2>/dev/null || true
./test.bqn

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_echo.bqn

echo ""
echo "=== Test: Argument passing ==="
./echo_args.bqn foo bar baz
