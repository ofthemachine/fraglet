#!/bin/sh
set -e

# Test JavaScript envelope by name
FRAGLETC="./fragletc"

echo "=== Test: Array processing with reduce ==="
cat <<'EOF' | "$FRAGLETC" --envelope javascript
const numbers = [1, 2, 3, 4, 5];
const squared = numbers.map(x => x * x);
const sum = squared.reduce((a, b) => a + b, 0);
console.log(`Sum of squares: ${sum}`);
EOF


