#!/bin/sh
set -e
chmod +x ./*.goz 2>/dev/null || true
./test.goz

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.goz

echo ""
echo "=== Test: Argument passing ==="
./echo_args.goz foo bar baz
