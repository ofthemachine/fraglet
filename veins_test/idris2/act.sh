#!/bin/sh
set -e
chmod +x ./*.idr 2>/dev/null || true
./test.idr
