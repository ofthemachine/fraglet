#!/bin/sh
set -e

# Test SNOBOL4 envelope by name
FRAGLETC="./fragletc"

echo "=== Test: Pattern matching and replacement ==="
cat <<'EOF' | "$FRAGLETC" --envelope snobol4
        TEXT = "Hello World"
        TEXT "World" = "Universe"
        OUTPUT = TEXT
END
EOF


