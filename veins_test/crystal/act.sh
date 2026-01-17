#!/bin/sh
set -e
chmod +x ./*.cr 2>/dev/null || true
./test.cr
