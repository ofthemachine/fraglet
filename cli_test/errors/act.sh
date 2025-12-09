#!/bin/sh
set -e

# Test fragletc error handling
FRAGLETC="./fragletc"

echo "=== Test 1: Missing required flag ==="
"$FRAGLETC" 2>&1 || true

echo ""
echo "=== Test 2: Both --image and --envelope (should error) ==="
echo 'print("test")' | "$FRAGLETC" --image 100hellos/python:latest --envelope python 2>&1 || true

echo ""
echo "=== Test 3: Invalid envelope name ==="
echo 'print("test")' | "$FRAGLETC" --envelope nonexistent 2>&1 || true

