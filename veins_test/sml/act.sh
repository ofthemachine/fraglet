#!/bin/sh
set -e
chmod +x ./*.sml 2>/dev/null || true
./test.sml

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.sml

echo ""
echo "=== Test: Argument passing ==="
./echo_args.sml foo bar baz
