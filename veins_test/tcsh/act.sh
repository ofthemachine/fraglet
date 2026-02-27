#!/bin/sh
set -e
chmod +x ./*.tcsh 2>/dev/null || true
./test.tcsh

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.tcsh

echo ""
echo "=== Test: Argument passing ==="
./echo_args.tcsh foo bar baz
