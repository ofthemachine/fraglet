#!/bin/sh
set -e
chmod +x ./*.js 2>/dev/null || true
./test.js

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.js

echo ""
echo "=== Test: Argument passing ==="
./echo_args.js foo bar baz
