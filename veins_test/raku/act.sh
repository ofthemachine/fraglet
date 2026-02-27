#!/bin/sh
set -e
chmod +x ./*.raku 2>/dev/null || true
./test.raku

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.raku

echo ""
echo "=== Test: Argument passing ==="
./echo_args.raku foo bar baz
