#!/bin/sh
set -e

# Test Python vein
# The fragletc binary is provided by the clitest harness with embedded veins
FRAGLETC="./fragletc"

echo "=== Test 1: Basic execution ==="
echo 'print("Hello, Python!")' | "$FRAGLETC" --vein python

echo ""
echo "=== Test 2: Multi-line code ==="
cat <<'EOF' | "$FRAGLETC" --vein python
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
"$FRAGLETC" --vein python test.py
rm -f test.py

echo ""
echo "=== Test 4: Extension inference ==="
cat > test_infer.py <<'EOF'
print("Inferred from .py extension!")
EOF
"$FRAGLETC" test_infer.py
rm -f test_infer.py
