#!/bin/sh
set -e
chmod +x ./*.clj 2>/dev/null || true
./test.clj

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.clj

echo ""
echo "=== Test: Argument passing ==="
./echo_args.clj foo bar baz
