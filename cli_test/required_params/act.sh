#!/bin/sh
set -e

# Required-param validation runs host-side, before any container is
# started, so these use -c/inline code with no --image/--vein at all --
# every case here must fail (or succeed) on argv parsing alone.

echo "=== Test 1: one required param, none supplied ==="
fragletc -c '# fraglet-meta: param=name:required
print("hi")' 2>&1 || true

echo ""
echo "=== Test 2: two required params, only one supplied ==="
fragletc -c '# fraglet-meta: param=name:required
# fraglet-meta: param=greeting:required
print("hi")' -p name=world 2>&1 || true

echo ""
echo "=== Test 3: all required params supplied -- validation passes, falls through to the next error (no image) ==="
fragletc -c '# fraglet-meta: param=name:required
print("hi")' -p name=world 2>&1 || true

echo ""
echo "=== Test 4: optional params are never required ==="
fragletc -c '# fraglet-meta: param=name:optional:default=world
print("hi")' 2>&1 || true

echo ""
echo "=== Test 5: an unknown -p alias surfaces its own error immediately, not a bogus missing-required one ==="
fragletc -c '# fraglet-meta: param=name:required
print("hi")' -p ghost=1 2>&1 || true

echo ""
echo "=== Test 6: required + default= together -- the default exempts it, not a contradiction ==="
fragletc -c '# fraglet-meta: param=name:required:default=world
print("hi")' 2>&1 || true
