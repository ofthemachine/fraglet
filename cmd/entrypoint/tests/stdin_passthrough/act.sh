#!/bin/sh
set -e

# Build the Docker image (binary is already in temp dir from harness)
docker build -t fraglet-stdin-test:latest -f Dockerfile . > /dev/null 2>&1

echo "=== Test 1: Basic stdin pipe ==="
echo "hello world" | docker run --rm -i -v "$(pwd)/fraglets/stdin-upper.sh:/FRAGLET:ro" fraglet-stdin-test:latest 2>&1

echo ""
echo "=== Test 2: Multi-line stdin ==="
printf "alpha\nbeta\ngamma\n" | docker run --rm -i -v "$(pwd)/fraglets/stdin-upper.sh:/FRAGLET:ro" fraglet-stdin-test:latest 2>&1

echo ""
echo "=== Test 3: Stdin + args ==="
echo "piped-data" | docker run --rm -i -v "$(pwd)/fraglets/stdin-args.sh:/FRAGLET:ro" fraglet-stdin-test:latest arg1 arg2 2>&1

echo ""
echo "=== Test 4: No stdin ==="
docker run --rm -v "$(pwd)/fraglets/no-stdin.sh:/FRAGLET:ro" fraglet-stdin-test:latest hello world 2>&1

echo ""
echo "=== Test 5: Byte count ==="
printf 'ABCDEFGHIJ' | docker run --rm -i -v "$(pwd)/fraglets/byte-count.sh:/FRAGLET:ro" fraglet-stdin-test:latest 2>&1

# Cleanup
docker rmi fraglet-stdin-test:latest > /dev/null 2>&1 || true
