#!/bin/sh
set -e

# Test Prolog vein by name
FRAGLETC="./fragletc"

echo "=== Test: Facts and queries ==="
cat <<'EOF' | "$FRAGLETC" --vein prolog
assertz(likes(alice, chocolate)).
assertz(likes(bob, ice_cream)).
likes(alice, What), write("Alice likes: "), write(What), nl.
halt.
EOF


