#!/bin/sh
set -e
chmod +x ./*.sno 2>/dev/null || true
./test.sno

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.sno

echo ""
echo "=== Test: Argument passing ==="
./echo_args.sno foo bar baz
