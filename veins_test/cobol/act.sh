#!/bin/sh
set -e
chmod +x ./*.cob 2>/dev/null || true
./test.cob

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.cob

echo ""
echo "=== Test: Argument passing ==="
./echo_args.cob foo bar baz
