#!/bin/sh
set -e

# Build a minimal test image with fraglet-entrypoint from GitHub releases,
# a default guide, and a mode-specific guide.
docker build --platform linux/amd64 -t fraglet-guide-test:local -f Dockerfile . > /dev/null 2>&1

# Point fragletc at our custom vein definition
export FRAGLET_VEINS_PATH=./veins.yml

echo "=== Test 1: Default guide (no mode) ==="
fragletc guide guidetester

echo ""
echo "=== Test 2: Mode flag BEFORE vein name (long form) ==="
fragletc guide --mode=testmode guidetester

echo ""
echo "=== Test 3: Mode flag AFTER vein name (long form, equals) ==="
fragletc guide guidetester --mode=testmode

echo ""
echo "=== Test 4: Mode flag AFTER vein name (long form, space) ==="
fragletc guide guidetester --mode testmode

echo ""
echo "=== Test 5: Short mode flag AFTER vein name ==="
fragletc guide guidetester -m testmode

echo ""
echo "=== Test 6: Short mode flag BEFORE vein name ==="
fragletc guide -m testmode guidetester

# Cleanup
docker rmi fraglet-guide-test:local > /dev/null 2>&1 || true
