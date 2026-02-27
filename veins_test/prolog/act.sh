#!/bin/sh
set -e
chmod +x ./*.pl 2>/dev/null || true
./test.pl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.pl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.pl foo bar baz
