#!/bin/sh
set -e

# Test LOLCODE vein by name
FRAGLETC="./fragletc"

echo "=== Test: Internet meme language ==="
cat <<'EOF' | "$FRAGLETC" --vein lolcode
VISIBLE "Fraglet"
VISIBLE "Rules!"
I HAS A VAR ITZ "awesome"
VISIBLE VAR
VISIBLE "Sum: "
VISIBLE SUM OF 10 AN 20
EOF


