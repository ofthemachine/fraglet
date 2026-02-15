#!/bin/sh
# Use local image so stdin works (vein defaults to :latest which may not have stdin support).
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.arnoldc 2>/dev/null || true
./test.arnoldc

echo ""
echo "=== Test: Stdin (integer read) ==="
echo "42" | ./stdin_int.arnoldc
