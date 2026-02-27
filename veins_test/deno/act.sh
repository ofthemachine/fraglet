#!/bin/sh
set -e
chmod +x ./*.ts 2>/dev/null || true
./test.ts

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ts

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ts foo bar baz
