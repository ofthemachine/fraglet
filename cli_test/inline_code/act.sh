#!/bin/sh
set -e

# Test fragletc -c (inline code) with direct container image

echo "=== Test 1: Inline code with --image flag ==="
fragletc -c 'print("Hello from inline!")' --image 100hellos/python:latest

echo ""
echo "=== Test 2: Inline code with short -i flag ==="
fragletc -c 'print("Hello with short flag!")' -i 100hellos/python:latest

echo ""
echo "=== Test 3: Inline code with custom --fraglet-path ==="
fragletc -c 'print("Custom path!")' --image 100hellos/python:latest --fraglet-path /FRAGLET

echo ""
echo "=== Test 4: Inline code with short -p flag ==="
fragletc -c 'print("Short path flag!")' -i 100hellos/python:latest -p /FRAGLET
