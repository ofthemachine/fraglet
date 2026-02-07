#!/bin/sh
set -e

# The binary and all test files are already in the temp test directory.
# Point the entrypoint at the test's fraglet.yaml (loader only checks FRAGLET_CONFIG or /fraglet.yaml).
export FRAGLET_CONFIG=fraglet.yaml
./fraglet-entrypoint

echo "---"
./fraglet-entrypoint usage
echo "---"
./fraglet-entrypoint guide
