#!/bin/sh
set -e
chmod +x ./*.gleam 2>/dev/null || true
./test.gleam
