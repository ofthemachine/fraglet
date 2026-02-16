#!/bin/sh
set -e
chmod +x ./*.scala 2>/dev/null || true
./test.scala

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.scala

echo ""
echo "=== Test: Argument passing ==="
./echo_args.scala foo bar baz
