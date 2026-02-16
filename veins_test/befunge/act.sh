#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
# No shebang in .bf (would be executed as Befunge code); run via fragletc
fragletc --vein=befunge test.bf

echo ""
echo "=== Test: Stdin ==="
echo "h" | fragletc --vein=befunge stdin_echo.bf
