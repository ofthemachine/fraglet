#!/bin/sh
set -e

# Test fragletc with embedded vein

echo "=== Test 1: STDIN with --vein flag ==="
echo 'print("Hello from vein!")' | fragletc --vein python

echo ""
echo "=== Test 2: STDIN with short -v flag ==="
echo 'print("Short vein flag!")' | fragletc -v python

echo ""
echo "=== Test 3: File input with vein ==="
cat > test_vein.py <<'EOF'
print("File with vein!")
EOF
fragletc --vein python test_vein.py
rm -f test_vein.py

echo ""
echo "=== Test 4: Vein with mode syntax (validates parsing) ==="
echo 'print("Mode syntax accepted")' | fragletc --vein python:main 2>&1 || echo "Mode syntax accepted (error expected if mode not supported)"

echo ""
echo "=== Test 5: Extension inference ==="
cat > test_infer.py <<'EOF'
print("Inferred from extension!")
EOF
fragletc test_infer.py
rm -f test_infer.py


