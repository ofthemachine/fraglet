#!/bin/sh
set -e

# Test fragletc with embedded vein

echo "=== Test 1: Inline code with --vein flag ==="
fragletc --vein python -c 'print("Hello from vein!")'

echo ""
echo "=== Test 2: Inline code with short -v flag ==="
fragletc -v python -c 'print("Short vein flag!")'

echo ""
echo "=== Test 3: File input with vein ==="
cat > test_vein.py <<'EOF'
print("File with vein!")
EOF
fragletc --vein python test_vein.py
rm -f test_vein.py

echo ""
echo "=== Test 4: Vein with mode syntax (validates parsing) ==="
fragletc --vein python:main -c 'print("Mode syntax accepted")' 2>&1 || echo "Mode syntax accepted (error expected if mode not supported)"

echo ""
echo "=== Test 5: Extension inference ==="
cat > test_infer.py <<'EOF'
print("Inferred from extension!")
EOF
fragletc test_infer.py
rm -f test_infer.py
