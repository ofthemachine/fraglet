#!/bin/sh
set -e
chmod +x ./*.java 2>/dev/null || true
./test.java

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.java

echo ""
echo "=== Test: Argument passing ==="
./echo_args.java foo bar baz
