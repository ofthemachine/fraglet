#!/bin/sh
set -e

# Test fragletc with embedded envelope by name
FRAGLETC="./fragletc"

echo "=== Test 1: STDIN with --envelope flag ==="
echo 'print("Hello from envelope!")' | "$FRAGLETC" --envelope python

echo ""
echo "=== Test 2: STDIN with short -e flag ==="
echo 'print("Short envelope flag!")' | "$FRAGLETC" -e python

echo ""
echo "=== Test 3: File input with envelope ==="
cat > test_envelope.py <<'EOF'
print("File with envelope!")
EOF
"$FRAGLETC" --envelope python --input test_envelope.py
rm -f test_envelope.py

echo ""
echo "=== Test 4: Envelope with custom fraglet-path ==="
echo 'print("Envelope custom path!")' | "$FRAGLETC" -e python --fraglet-path /FRAGLET


