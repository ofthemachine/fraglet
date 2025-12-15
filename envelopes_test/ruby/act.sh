#!/bin/sh
set -e

# Test Ruby envelope by name
FRAGLETC="./fragletc"

echo "=== Test: Array processing ==="
cat <<'EOF' | "$FRAGLETC" --envelope ruby
numbers = [1, 2, 3, 4, 5]
squared = numbers.map { |x| x**2 }
puts "Sum of squares: #{squared.sum}"
EOF


