#!/bin/sh
set -e
chmod +x ./*.st 2>/dev/null || true
./test.st

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.st

echo ""
echo "=== Test: Argument passing ==="
./echo_args.st foo bar baz
