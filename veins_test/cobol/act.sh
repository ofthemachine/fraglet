#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.cob 2>/dev/null || true
./test.cob

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.cob

echo ""
echo "=== Test: Argument passing ==="
./echo_args.cob foo bar baz
./echo_args.cob "queen foo" 1 2

echo ""
echo "=== Test: COBOL fortune ==="
./cobol_fortune.cob
