#!/bin/sh
set -e
chmod +x ./*.pony 2>/dev/null || true
./test.pony

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.pony

echo ""
echo "=== Test: Argument passing ==="
./echo_args.pony foo bar baz
