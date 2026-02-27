#!/bin/sh
set -e
chmod +x ./*.pas 2>/dev/null || true
./test.pas

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.pas

echo ""
echo "=== Test: Argument passing ==="
./echo_args.pas foo bar baz
