#!/bin/sh
set -e
chmod +x ./*.nim 2>/dev/null || true
./test.nim
