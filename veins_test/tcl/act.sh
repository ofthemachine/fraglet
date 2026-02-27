#!/bin/sh
set -e
chmod +x ./*.tcl 2>/dev/null || true
./test.tcl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.tcl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.tcl foo bar baz
