#!/bin/sh
set -e
chmod +x ./*.py 2>/dev/null || true
./test.py

echo ""
echo "=== Test: Argument passing ==="
./echo_args.py foo bar baz
