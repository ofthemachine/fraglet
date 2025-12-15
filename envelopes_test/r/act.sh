#!/bin/sh
set -e

# Test R envelope by name
FRAGLETC="./fragletc"

echo "=== Test: Vector processing ==="
cat <<'EOF' | "$FRAGLETC" --envelope r
numbers <- 1:5
squared <- numbers^2
cat("Sum of squares:", sum(squared), "\n")
EOF


