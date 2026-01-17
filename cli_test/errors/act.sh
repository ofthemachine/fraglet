#!/bin/sh
set -e

# Test fragletc error handling
FRAGLETC="./fragletc"

echo "=== Test 1: Missing required flag ==="
"$FRAGLETC" 2>&1 || true

echo ""
echo "=== Test 2: Both --image and --vein (should error) ==="
echo 'print("test")' | "$FRAGLETC" --image 100hellos/python:latest --vein python 2>&1 || true

echo ""
echo "=== Test 3: Invalid vein name ==="
echo 'print("test")' | "$FRAGLETC" --vein nonexistent 2>&1 || true

