#!/bin/sh
set -e

# Test Julia envelope by name
FRAGLETC="./fragletc"

echo "=== Test: High-performance numerical computing ==="
cat <<'EOF' | "$FRAGLETC" --envelope julia
arr = [1, 2, 3, 4, 5]
sum_val = sum(arr)
println("Array: ", arr)
println("Sum: ", sum_val)
println("Product: ", prod(arr))
EOF

