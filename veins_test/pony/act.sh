#!/bin/sh
set -e
chmod +x ./*.pony 2>/dev/null || true
./test.pony
