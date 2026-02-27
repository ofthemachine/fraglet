#!/bin/sh
set -e
chmod +x ./*.rkt 2>/dev/null || true
./test.rkt

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.rkt

echo ""
echo "=== Test: Argument passing ==="
./echo_args.rkt foo bar baz
