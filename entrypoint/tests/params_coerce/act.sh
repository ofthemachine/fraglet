#!/bin/sh
set -e

# Build the Docker image (binary is already in temp dir from harness)
docker build -t fraglet-params-test:latest -f Dockerfile . > /dev/null 2>&1

echo "=== Test 1: Basic param injection ==="
docker run --rm \
  -e FRAGLET_PARAM_CITY=london \
  -v "$(pwd)/fraglets/print_city.sh:/FRAGLET:ro" \
  fraglet-params-test:latest 2>&1

echo ""
echo "=== Test 2: Multiple params ==="
docker run --rm \
  -e FRAGLET_PARAM_CITY=paris \
  -e FRAGLET_PARAM_UNITS=metric \
  -v "$(pwd)/fraglets/print_multi.sh:/FRAGLET:ro" \
  fraglet-params-test:latest 2>&1

echo ""
echo "=== Test 3: Case-preserved env var name ==="
docker run --rm \
  -e FRAGLET_PARAM_HURL_VARIABLE_host=localhost \
  -v "$(pwd)/fraglets/print_mixed_case.sh:/FRAGLET:ro" \
  fraglet-params-test:latest 2>&1

echo ""
echo "=== Test 4: No-shadow (HOME already exists) ==="
docker run --rm \
  -e FRAGLET_PARAM_HOME=injected \
  -v "$(pwd)/fraglets/check_no_shadow.sh:/FRAGLET:ro" \
  fraglet-params-test:latest 2>&1

echo ""
echo "=== Test 5: Transport var cleaned after coercion ==="
docker run --rm \
  -e FRAGLET_PARAM_CITY=tokyo \
  -v "$(pwd)/fraglets/check_transport_cleaned.sh:/FRAGLET:ro" \
  fraglet-params-test:latest 2>&1

echo ""
echo "=== Test 6: No params — runs normally ==="
docker run --rm \
  -v "$(pwd)/fraglets/no_params.sh:/FRAGLET:ro" \
  fraglet-params-test:latest 2>&1

# Cleanup
docker rmi fraglet-params-test:latest > /dev/null 2>&1 || true
