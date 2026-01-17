#!/bin/sh
set -e

# Test fragletc with STDIN input using direct container image

echo "=== Test 1: STDIN with --image flag ==="
echo 'print("Hello from STDIN!")' | fragletc --image 100hellos/python:latest

echo ""
echo "=== Test 2: STDIN with short -i flag ==="
echo 'print("Hello with short flag!")' | fragletc -i 100hellos/python:latest

echo ""
echo "=== Test 3: STDIN with custom --fraglet-path ==="
echo 'print("Custom path!")' | fragletc --image 100hellos/python:latest --fraglet-path /FRAGLET

echo ""
echo "=== Test 4: STDIN with short -p flag ==="
echo 'print("Short path flag!")' | fragletc -i 100hellos/python:latest -p /FRAGLET

