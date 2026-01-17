#!/bin/sh
set -e

# Test fragletc error handling

echo "=== Test 1: Missing required flag ==="
fragletc 2>&1 || true

echo ""
echo "=== Test 2: Both --image and --vein (should error) ==="
echo 'print("test")' | fragletc --image 100hellos/python:latest --vein python 2>&1 || true

echo ""
echo "=== Test 3: Invalid vein name ==="
echo 'print("test")' | fragletc --vein nonexistent 2>&1 || true

