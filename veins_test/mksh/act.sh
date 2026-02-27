#!/bin/sh
set -e
chmod +x ./*.mksh 2>/dev/null || true
./test.mksh

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.mksh

echo ""
echo "=== Test: Argument passing ==="
./echo_args.mksh foo bar baz
