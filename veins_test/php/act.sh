#!/bin/sh
set -e
chmod +x ./*.php 2>/dev/null || true
./test.php

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.php

echo ""
echo "=== Test: Argument passing ==="
./echo_args.php foo bar baz
