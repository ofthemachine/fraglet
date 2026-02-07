#!/bin/sh
set -e

# Test fragletc error handling

echo "=== Test 1: No args (should show usage) ==="
fragletc 2>&1 || true

echo ""
echo "=== Test 2: Both --image and --vein (should error) ==="
fragletc -c 'print("test")' --image 100hellos/python:latest --vein python 2>&1 || true

echo ""
echo "=== Test 3: Invalid vein name ==="
fragletc -c 'print("test")' --vein nonexistent 2>&1 || true

echo ""
echo "=== Test 4: No code source with vein ==="
fragletc --vein python 2>&1 || true
