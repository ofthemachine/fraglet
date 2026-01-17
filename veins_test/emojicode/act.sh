#!/bin/sh
set -e

# Test EmojiCode vein by name
FRAGLETC="./fragletc"

echo "=== Test: Multiple outputs and conditionals ==="
cat <<'EOF' | "$FRAGLETC" --vein emojicode
😀 🔤Fraglet Test🔤❗️
😀 🔤Multiple lines🔤❗️
😀 🔤of output🔤❗️
↪️ 👍 🍇
  😀 🔤Condition is true!🔤❗️
🍉
🙅 🍇
  😀 🔤This won't print🔤❗️
🍉
EOF
