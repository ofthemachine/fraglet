#!/bin/sh
set -e
chmod +x ./*.awk 2>/dev/null || true
./test.awk
