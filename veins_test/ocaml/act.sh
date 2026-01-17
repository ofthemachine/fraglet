#!/bin/sh
set -e
chmod +x ./*.ml 2>/dev/null || true
./test.ml
