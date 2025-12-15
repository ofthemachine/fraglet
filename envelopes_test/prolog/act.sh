#!/bin/sh
set -e

# Test Prolog envelope by name
FRAGLETC="./fragletc"

echo "=== Test: Facts and queries ==="
cat <<'EOF' | "$FRAGLETC" --envelope prolog
assertz(likes(alice, chocolate)).
assertz(likes(bob, ice_cream)).
likes(alice, What), write("Alice likes: "), write(What), nl.
halt.
EOF


