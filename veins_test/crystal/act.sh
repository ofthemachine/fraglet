#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.cr 2>/dev/null || true
./test.cr

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_echo.cr

echo ""
echo "=== Test: Argument passing ==="
./echo_args.cr foo bar baz
