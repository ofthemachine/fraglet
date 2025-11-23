#!/bin/sh
set -e

# The binary and all test files are already in the temp test directory
# The fraglet.yaml is already a sibling to the binary (both in temp dir root)
./fraglet-entrypoint

echo "---"
./fraglet-entrypoint usage
echo "---"
./fraglet-entrypoint guide
