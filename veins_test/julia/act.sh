#!/bin/sh
set -e
chmod +x ./*.jl 2>/dev/null || true
./test.jl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.jl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.jl foo bar baz
