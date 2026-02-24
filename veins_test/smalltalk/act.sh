#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.st 2>/dev/null || true
./test.st

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.st

echo ""
echo "=== Test: Argument passing ==="
./echo_args.st foo bar baz
./echo_args.st "queen foo" 1 2

echo ""
echo "=== Test: Smalltalk fortune ==="
./smalltalk_fortune.st
