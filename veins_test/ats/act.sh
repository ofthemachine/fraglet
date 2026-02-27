#!/bin/sh
set -e
chmod +x ./*.dats 2>/dev/null || true
./test.dats

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.dats

echo ""
echo "=== Test: Argument passing ==="
./echo_args.dats foo bar baz
