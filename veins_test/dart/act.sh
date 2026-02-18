#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.dart 2>/dev/null || true
./test.dart

echo ""
echo "=== Test: Stdin (shout back) ==="
echo "hello" | ./stdin_echo.dart

echo ""
echo "=== Test: Argument passing ==="
./echo_args.dart foo bar baz
