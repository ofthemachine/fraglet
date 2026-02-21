#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.rs 2>/dev/null || true
./test.rs

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.rs

echo ""
echo "=== Test: Argument passing ==="
./echo_args.rs foo bar baz
./echo_args.rs "queen foo" 1 2

echo ""
echo "=== Test: Rust fortune ==="
./rust_fortune.rs
