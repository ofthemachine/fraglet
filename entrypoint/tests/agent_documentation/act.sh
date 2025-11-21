#!/bin/sh
mkdir -p code code-fragments

export FRAGLET_CONFIG=${PWD}/fraglet-entrypoint.yaml
fraglet-entrypoint agent-help
fraglet-entrypoint how-to

rm fraglet/agent-help.md
fraglet-entrypoint agent-help

rm fraglet/how-to.md
fraglet-entrypoint how-to
