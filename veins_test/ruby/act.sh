#!/bin/sh
set -e

# Test Ruby vein by name
FRAGLETC="./fragletc"

echo "=== Test: Array processing ==="
cat <<'EOF' | "$FRAGLETC" --vein ruby
numbers = [1, 2, 3, 4, 5]
squared = numbers.map { |x| x**2 }
puts "Sum of squares: #{squared.sum}"
EOF


