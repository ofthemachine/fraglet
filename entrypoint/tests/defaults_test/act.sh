#!/bin/sh
set -e

# The binary and all test files are already in the temp test directory
# The fraglet-entrypoint.yaml is already a sibling to the binary (both in temp dir root)
# So we just need to unset the envvar and run the commands

unset FRAGLET_CONFIG
./fraglet-entrypoint

echo "---"
./fraglet-entrypoint agent-help
echo "---"
./fraglet-entrypoint how-to
