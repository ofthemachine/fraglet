#!/bin/sh
set -e
chmod +x ./*.d 2>/dev/null || true
./test.d
