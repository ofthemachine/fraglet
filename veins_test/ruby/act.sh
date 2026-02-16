#!/bin/sh
set -e
chmod +x ./*.rb 2>/dev/null || true
./test.rb

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.rb

echo ""
echo "=== Test: Argument passing ==="
./echo_args.rb foo bar baz
