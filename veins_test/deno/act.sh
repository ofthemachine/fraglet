#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.ts 2>/dev/null || true
./test.ts

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.ts

echo ""
echo "=== Test: Argument passing ==="
./echo_args.ts foo bar baz

echo ""
echo "=== Test: Deno fortune ==="
./deno_fortune.ts
