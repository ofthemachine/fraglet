#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.groovy 2>/dev/null || true
./test.groovy

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.groovy

echo ""
echo "=== Test: Argument passing ==="
./echo_args.groovy foo bar baz

echo ""
echo "=== Test: Groovy fortune ==="
./groovy_fortune.groovy
