#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.dash 2>/dev/null || true
./test.dash

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.dash

echo ""
echo "=== Test: Argument passing ==="
./echo_args.dash foo bar baz

echo ""
echo "=== Test: Dash fortune ==="
./dash_fortune.dash
