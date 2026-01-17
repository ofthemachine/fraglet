#!/bin/sh
set -e

# Test ArnoldC vein by name
FRAGLETC="./fragletc"

echo "=== Test: Arnold Schwarzenegger quotes as code ==="
cat <<'EOF' | "$FRAGLETC" --vein=arnoldc
TALK TO THE HAND "Fraglet"
TALK TO THE HAND "Rules!"
EOF
