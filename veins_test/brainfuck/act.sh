#!/bin/sh
set -e
chmod +x ./*.bf 2>/dev/null || true
./test.bf

echo ""
echo "=== Test: Stdin (one char) ==="
printf "h" | ./stdin_echo.bf
