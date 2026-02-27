#!/bin/sh
set -e
fragletc --vein=befunge test.bf

echo ""
echo "=== Factorial 5! ==="
fragletc --vein=befunge factorial.bf

echo ""
echo "=== Squares 1..5 ==="
fragletc --vein=befunge squares.bf

echo ""
echo "=== Test: Stdin ==="
echo "h" | fragletc --vein=befunge stdin_echo.bf
