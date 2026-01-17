#!/bin/sh
set -e

# Test TypeScript vein by name
FRAGLETC="./fragletc"

echo "=== Test: Typed array processing ==="
cat <<'EOF' | "$FRAGLETC" --vein typescript
const numbers: number[] = [1, 2, 3, 4, 5];
const squared: number[] = numbers.map((x: number) => x * x);
const sum: number = squared.reduce((a: number, b: number) => a + b, 0);
console.log(`Sum of squares: ${sum}`);
EOF


