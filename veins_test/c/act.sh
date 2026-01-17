#!/bin/sh
set -e
chmod +x ./*.c 2>/dev/null || true

echo "=== Test: c ==="
./array_sum.c


