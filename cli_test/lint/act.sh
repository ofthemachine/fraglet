#!/bin/sh
# fragletc lint: header grammar errors, convention warnings, --strict, and
# directory walking (fragletc-shebang files only, hidden dirs skipped).

echo "=== Test 1: clean file exits 0 ==="
fragletc lint clean.py
echo "exit=$?"

echo ""
echo "=== Test 2: offenders report every rule, exit 1 ==="
fragletc lint offenders.py
echo "exit=$?"

echo ""
echo "=== Test 3: warnings alone exit 0 ==="
fragletc lint tree/warn_only.py
echo "exit=$?"

echo ""
echo "=== Test 4: --strict turns warnings into errors, exit 1 ==="
fragletc lint --strict tree/warn_only.py
echo "exit=$?"

echo ""
echo "=== Test 5: directory walk lints fragletc-shebang files only ==="
fragletc lint tree
echo "exit=$?"

echo ""
echo "=== Test 6: an explicitly named non-fragletc file is still linted ==="
fragletc lint tree/not_fragletc.sh
echo "exit=$?"

echo ""
echo "=== Test 7: no paths is a usage error, exit 2 ==="
fragletc lint 2>/dev/null
echo "exit=$?"

echo ""
echo "=== Test 8: missing path, exit 2 ==="
fragletc lint does-not-exist.py 2>&1
echo "exit=$?"
