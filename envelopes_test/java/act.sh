#!/bin/sh
set -e

# Test Java envelope by name
FRAGLETC="./fragletc"

echo "=== Test: WordSet operations ==="
cat <<'EOF' | "$FRAGLETC" --envelope java
WordSet<?> words = HelloWorld.loadWords();
int count = words.endingWith("ing").count();
System.out.println("Words ending with 'ing': " + count);
EOF

