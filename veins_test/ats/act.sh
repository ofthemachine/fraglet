#!/bin/sh
set -e
chmod +x ./*.dats 2>/dev/null || true
./test.dats
