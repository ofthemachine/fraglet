#!/bin/sh
set -e

# Test SNOBOL4 vein by name
FRAGLETC="./fragletc"

echo "=== Test: Pattern matching and replacement ==="
cat <<'EOF' | "$FRAGLETC" --vein snobol4
        TEXT = "Hello World"
        TEXT "World" = "Universe"
        OUTPUT = TEXT
END
EOF


