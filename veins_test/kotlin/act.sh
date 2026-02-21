#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.kt 2>/dev/null || true
./test.kt

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.kt

echo ""
echo "=== Test: Argument passing ==="
./echo_args.kt foo bar baz
./echo_args.kt "queen foo" 1 2

echo ""
echo "=== Test: Kotlin fortune ==="
./kotlin_fortune.kt
