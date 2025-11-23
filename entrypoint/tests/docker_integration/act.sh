#!/bin/sh
set -e

# Build the Docker image (binary is already in temp dir from harness)
docker build -t fraglet-test:latest -f Dockerfile . > /dev/null 2>&1

# Create a temporary fraglet file
echo 'echo "🎉 Fraglet injected successfully!"' > /tmp/test-fraglet.sh

# Test 1: Single match substitution (default config)
echo "=== Test 1: Single match ==="
docker run --rm -v /tmp/test-fraglet.sh:/FRAGLET:ro fraglet-test:latest 2>&1

# Test 2: Range-based substitution (alternative config)
echo ""
echo "=== Test 2: Range-based match ==="
docker run --rm \
  -v /tmp/test-fraglet.sh:/FRAGLET:ro \
  -e FRAGLET_CONFIG=/fraglet-alternative.yaml \
  fraglet-test:latest 2>&1

# Test the usage command
echo ""
echo "---"
echo "Testing usage:"
docker run --rm fraglet-test:latest usage

echo ""
echo "---"
echo "Testing guide:"
docker run --rm fraglet-test:latest guide

# Cleanup
docker rmi fraglet-test:latest > /dev/null 2>&1 || true
