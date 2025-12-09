#!/bin/sh
set -e

# Test fragletc with file input
FRAGLETC="./fragletc"

echo "=== Test 1: File input with --input flag ==="
cat > test_input.py <<'EOF'
print("Hello from file!")
EOF
"$FRAGLETC" --image 100hellos/python:latest --input test_input.py
rm -f test_input.py

echo ""
echo "=== Test 2: File input with short -f flag ==="
cat > test_input2.py <<'EOF'
for i in range(3):
    print(f"Count: {i}")
EOF
"$FRAGLETC" -i 100hellos/python:latest -f test_input2.py
rm -f test_input2.py

echo ""
echo "=== Test 3: File input with custom fraglet-path ==="
cat > test_input3.py <<'EOF'
print("File with custom path!")
EOF
"$FRAGLETC" --image 100hellos/python:latest --fraglet-path /FRAGLET --input test_input3.py
rm -f test_input3.py

