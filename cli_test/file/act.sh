#!/bin/sh
set -e

# Test fragletc with file input

echo "=== Test 1: File input with --image flag ==="
cat > test_input.py <<'EOF'
print("Hello from file!")
EOF
fragletc --image 100hellos/python:latest test_input.py
rm -f test_input.py

echo ""
echo "=== Test 2: File input with short -i flag ==="
cat > test_input2.py <<'EOF'
for i in range(3):
    print(f"Count: {i}")
EOF
fragletc -i 100hellos/python:latest test_input2.py
rm -f test_input2.py

echo ""
echo "=== Test 3: File input with custom fraglet-path ==="
cat > test_input3.py <<'EOF'
print("File with custom path!")
EOF
fragletc --image 100hellos/python:latest --fraglet-path /FRAGLET test_input3.py
rm -f test_input3.py

