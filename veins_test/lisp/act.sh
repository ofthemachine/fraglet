#!/bin/sh
set -e
chmod +x ./*.lisp 2>/dev/null || true
./test.lisp

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.lisp

echo ""
echo "=== Test: Argument passing ==="
./echo_args.lisp foo bar baz
