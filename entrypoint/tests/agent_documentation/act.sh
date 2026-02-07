#!/bin/sh
# Load test's fraglet.yaml (loader only checks FRAGLET_CONFIG or /fraglet.yaml)
export FRAGLET_CONFIG=fraglet.yaml

fraglet-entrypoint usage
fraglet-entrypoint guide

rm fraglet/guide.md
fraglet-entrypoint guide
