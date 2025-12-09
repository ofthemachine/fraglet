#!/bin/sh
set -e

# Test fragletc with Python envelope by name
# The fragletc binary is provided by the clitest harness in the test directory
FRAGLETC="./fragletc"

echo "=== Test 1: Simple print from STDIN using envelope name ==="
echo 'print("Hello from STDIN!")' | "$FRAGLETC" --envelope python

echo ""
echo "=== Test 2: Simple print with explicit fraglet-path ==="
echo 'print("Hello with explicit path!")' | "$FRAGLETC" --envelope python --fraglet-path /FRAGLET

echo ""
echo "=== Test 3: Multi-line code ==="
cat <<'EOF' | "$FRAGLETC" --envelope python
def greet(name):
    return f"Hello, {name}!"
print(greet("Fraglet"))
EOF

echo ""
echo "=== Test 4: Code from file ==="
cat > test_python.py <<'EOF'
for i in range(3):
    print(f"Count: {i}")
EOF
"$FRAGLETC" --envelope python --input test_python.py
rm -f test_python.py
