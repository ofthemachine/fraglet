#!/bin/sh
set -e

# Test Python envelope by name
# The fragletc binary is provided by the clitest harness with embedded envelopes
FRAGLETC="./fragletc"

echo "=== Test 1: Basic execution ==="
echo 'print("Hello, Python!")' | "$FRAGLETC" --envelope python

echo ""
echo "=== Test 2: Multi-line code ==="
cat <<'EOF' | "$FRAGLETC" --envelope python
def greet(name):
    return f"Hello, {name}!"
print(greet("Fraglet"))
EOF

echo ""
echo "=== Test 3: File input ==="
cat > test.py <<'EOF'
for i in range(3):
    print(f"Count: {i}")
EOF
"$FRAGLETC" --envelope python --input test.py
rm -f test.py
