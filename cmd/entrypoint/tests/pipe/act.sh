#!/bin/sh
set -e

# Test: Pipe data through Python -> Ruby -> C
# Each language adds its own prefix and processes the data
# This demonstrates stdin passthrough across multiple fraglet scripts
# The shebangs make them executable directly - no docker run needed!
chmod +x files/code/* 2>/dev/null || true

echo "hello world" | \
  ./files/code/python-processor.py | \
  ./files/code/ruby-processor.rb | \
  ./files/code/c-processor.c