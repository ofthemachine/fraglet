#!/bin/sh
set -e
chmod +x ./*.R 2>/dev/null || true

echo "=== Test: Vector processing ==="
./vector_process.R


