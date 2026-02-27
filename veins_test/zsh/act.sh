#!/bin/sh
set -e
chmod +x ./*.zsh 2>/dev/null || true
./test.zsh

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.zsh

echo ""
echo "=== Test: Argument passing ==="
./echo_args.zsh foo bar baz
