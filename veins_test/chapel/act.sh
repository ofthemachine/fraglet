#!/bin/sh
set -e
chmod +x ./*.chpl 2>/dev/null || true
./test.chpl
