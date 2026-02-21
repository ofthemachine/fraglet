#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.rkt 2>/dev/null || true
./test.rkt

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.rkt

echo ""
echo "=== Test: Argument passing ==="
./echo_args.rkt foo bar baz
./echo_args.rkt "queen foo" 1 2

echo ""
echo "=== Test: Racket fortune ==="
./racket_fortune.rkt
