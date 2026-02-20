#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.tcsh 2>/dev/null || true
./test.tcsh

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.tcsh

echo ""
echo "=== Test: Argument passing ==="
./echo_args.tcsh foo bar baz

echo ""
echo "=== Test: tcsh fortune ==="
./tcsh_fortune.tcsh
