#!/bin/sh

fraglet-entrypoint usage
fraglet-entrypoint guide

rm fraglet/guide.md
fraglet-entrypoint guide
