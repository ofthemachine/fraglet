#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.tcl 2>/dev/null || true
./test.tcl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.tcl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.tcl foo bar baz

echo ""
echo "=== Test: Tcl fortune ==="
./tcl_fortune.tcl
