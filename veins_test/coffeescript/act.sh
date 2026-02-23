#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.coffee 2>/dev/null || true
./test.coffee

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.coffee

echo ""
echo "=== Test: Argument passing ==="
./echo_args.coffee foo bar baz
./echo_args.coffee "queen foo" 1 2

echo ""
echo "=== Test: CoffeeScript fortune ==="
./coffeescript_fortune.coffee
