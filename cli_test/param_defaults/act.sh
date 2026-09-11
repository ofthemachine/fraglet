#!/bin/sh
set -e

# default= is injected into the container env when -p is omitted.
# Explicit -p always wins (including empty string).

echo "=== Test 1: omitted -p gets declared default ==="
fragletc --vein python -c '#: param=encoding:default=cl100k_base
import os
print(os.environ.get("ENCODING", "UNSET"))'

echo ""
echo "=== Test 2: explicit -p overrides default ==="
fragletc --vein python -c '#: param=encoding:default=cl100k_base
import os
print(os.environ.get("ENCODING", "UNSET"))' -p encoding=o200k_base

echo ""
echo "=== Test 3: description= appears in --fraglet-help ==="
fragletc --fraglet-help -c '#: param=settle_ms:default=2000:description=Extra wait after page load
#: param=url:required:d=Target page URL
print(1)'
