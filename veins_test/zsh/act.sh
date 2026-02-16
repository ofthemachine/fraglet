#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.zsh 2>/dev/null || true
./test.zsh

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_echo.zsh

echo ""
echo "=== Test: Argument passing ==="
./echo_args.zsh foo bar baz
