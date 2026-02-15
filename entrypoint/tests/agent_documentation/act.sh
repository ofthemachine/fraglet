#!/bin/sh
set -e
# Binary is built for linux/amd64; run it inside Docker so it works on any host.
export FRAGLET_CONFIG=fraglet.yaml
run() { docker run --rm --platform linux/amd64 -v "$(pwd):/work" -w /work -e FRAGLET_CONFIG=fraglet.yaml alpine:latest /work/fraglet-entrypoint "$@"; }

run usage
run guide

rm -f fraglet/guide.md
run guide
