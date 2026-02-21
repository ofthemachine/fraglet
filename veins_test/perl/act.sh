#!/bin/sh
export FRAGLET_VEINS_FORCE_TAG=local
set -e
chmod +x ./*.pl 2>/dev/null || true
./test.pl

echo ""
echo "=== Test: Stdin ==="
echo "hello" | ./stdin_upper.pl

echo ""
echo "=== Test: Argument passing ==="
./echo_args.pl foo bar baz
./echo_args.pl "queen foo" 1 2

echo ""
echo "=== Test: Perl fortune ==="
./perl_fortune.pl
