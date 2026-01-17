#!/bin/sh
set -e
chmod +x ./*.cpp 2>/dev/null || true
./test.cpp

echo ""
echo "=== Test: Argument passing ==="
./echo_args.cpp foo bar baz
