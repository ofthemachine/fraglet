#!/bin/sh
set -e

# Test C programming language envelopes by name
FRAGLETC="./fragletc"

echo "=== Test: the-c-programming-language ==="
cat <<'EOF' | "$FRAGLETC" --envelope the-c-programming-language
int numbers[] = {1, 2, 3, 4, 5};
int sum = 0;
for (int i = 0; i < 5; i++) {
    sum += numbers[i];
}
printf("Array sum: %d\n", sum);
EOF

echo ""
echo "=== Test: the-c-programming-language-main ==="
cat <<'EOF' | "$FRAGLETC" --envelope the-c-programming-language-main
int a = 10;
int b = 20;
printf("Sum: %d\n", a + b);
printf("Product: %d\n", a * b);
EOF


