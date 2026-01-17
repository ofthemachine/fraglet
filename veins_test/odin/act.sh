#!/bin/sh
set -e
chmod +x ./*.odin 2>/dev/null || true
./test.odin
