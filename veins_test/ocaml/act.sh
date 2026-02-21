#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.ml 2>/dev/null || true
./test.ml

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ml

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ml foo bar baz
./echo_args.ml "queen foo" 1 2

echo ""
echo "=== Test: OCaml fortune ==="
./ocaml_fortune.ml
