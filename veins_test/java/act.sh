#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.java 2>/dev/null || true
./test.java

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.java

echo ""
echo "=== Test: Argument passing ==="
./echo_args.java foo bar baz
./echo_args.java "queen foo" 1 2

echo ""
echo "=== Test: Java fortune ==="
./java_fortune.java

echo ""
echo "=== Test: Wordalytica mode (shebang) ==="
./wordalytica_ing_count.java
