#!/bin/sh
set -e

# Test fragletc with embedded vein
FRAGLETC="./fragletc"

echo "=== Test 1: STDIN with --vein flag ==="
echo 'print("Hello from vein!")' | "$FRAGLETC" --vein python

echo ""
echo "=== Test 2: STDIN with short -v flag ==="
echo 'print("Short vein flag!")' | "$FRAGLETC" -v python

echo ""
echo "=== Test 3: File input with vein ==="
cat > test_vein.py <<'EOF'
print("File with vein!")
EOF
"$FRAGLETC" --vein python test_vein.py
rm -f test_vein.py

echo ""
echo "=== Test 4: Vein with mode ==="
echo 'print("Vein with mode!")' | "$FRAGLETC" --vein python:main

echo ""
echo "=== Test 5: Extension inference ==="
cat > test_infer.py <<'EOF'
print("Inferred from extension!")
EOF
"$FRAGLETC" test_infer.py
rm -f test_infer.py


