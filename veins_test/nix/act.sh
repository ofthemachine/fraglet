#!/bin/sh
set -e
chmod +x ./*.nix 2>/dev/null || true
./test.nix

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.nix

echo ""
echo "=== Test: Argument passing ==="
./echo_args.nix foo bar baz
