#!/bin/sh
set -e
chmod +x ./*.fnl 2>/dev/null || true
./test.fnl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.fnl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.fnl foo bar baz
