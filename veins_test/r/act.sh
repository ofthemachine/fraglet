#!/bin/sh
set -e

# Test R vein by name
FRAGLETC="./fragletc"

echo "=== Test: Vector processing ==="
cat <<'EOF' | "$FRAGLETC" --vein r
numbers <- 1:5
squared <- numbers^2
cat("Sum of squares:", sum(squared), "\n")
EOF


