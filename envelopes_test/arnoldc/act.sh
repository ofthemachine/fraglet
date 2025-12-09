#!/bin/sh
set -e

# Test ArnoldC envelope by name
FRAGLETC="./fragletc"

echo "=== Test: Arnold Schwarzenegger quotes as code ==="
cat <<'EOF' | "$FRAGLETC" --envelope arnoldc
TALK TO THE HAND "Fraglet"
TALK TO THE HAND "Rules!"
EOF
