#!/bin/sh
set -e
chmod +x ./*.ha 2>/dev/null || true
./test.ha

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ha

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ha foo bar baz
