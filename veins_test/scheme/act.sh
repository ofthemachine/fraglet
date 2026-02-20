#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.scm 2>/dev/null || true
./test.scm

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.scm

echo ""
echo "=== Test: Argument passing ==="
./echo_args.scm foo bar baz

echo ""
echo "=== Test: Scheme fortune ==="
./scheme_fortune.scm
