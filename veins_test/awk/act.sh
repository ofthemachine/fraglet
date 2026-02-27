#!/bin/sh
set -e
chmod +x ./*.awk 2>/dev/null || true
./test.awk

echo ""
echo "=== Test: Argument passing ==="
./echo_args.awk foo bar baz
