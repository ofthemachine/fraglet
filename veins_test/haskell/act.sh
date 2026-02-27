#!/bin/sh
set -e
chmod +x ./*.hs 2>/dev/null || true
./test.hs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.hs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.hs foo bar baz
