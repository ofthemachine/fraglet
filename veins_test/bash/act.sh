#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.bash 2>/dev/null || true
./test.bash

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.bash

echo ""
echo "=== Test: Argument passing ==="
./echo_args.bash foo bar baz
