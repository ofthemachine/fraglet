#!/bin/sh
set -e
chmod +x ./*.factor 2>/dev/null || true
./test.factor
