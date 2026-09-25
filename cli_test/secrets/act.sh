#!/bin/sh
set -e

# #: secret=NAME -- the value comes from the caller's env var NAME and
# reaches the program only as the read-only file $NAME_FILE
# (/run/fraglet/secrets/NAME). It is never a container env var, never on
# the docker command line, and never in a receipt.

CODE='#: secret=FT_TOKEN:d=Test credential
import os
path = os.environ["FT_TOKEN_FILE"]
value = open(path).read()
print("path:", path)
print("value:", value)
print("as env var:", "FT_TOKEN" in os.environ)
print("value anywhere in environ:", any(value in v for v in os.environ.values()))'

echo "=== Test 1: missing secret fails host-side (exit 2) ==="
unset FT_TOKEN
fragletc --vein python -c "$CODE" 2>&1 || echo "exit=$?"

echo ""
echo "=== Test 2: delivered as a file, never as an env var (multi-line ok) ==="
export FT_TOKEN='line-one
line-two'
fragletc --vein python -c "$CODE"

echo ""
echo "=== Test 3: forwarding a declared secret with -e is refused ==="
fragletc --vein python -e FT_TOKEN -c "$CODE" 2>&1 || echo "exit=$?"

echo ""
echo "=== Test 4: the receipt never carries the value, and the run keeps its memo key ==="
export FT_TOKEN='receipt-sentinel-value'
fragletc --vein python --receipt secret-receipt.json -c "$CODE" >/dev/null 2>&1
grep -c receipt-sentinel-value secret-receipt.json || true
grep -c '"memo_key"' secret-receipt.json
