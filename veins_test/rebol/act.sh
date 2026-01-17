#!/bin/sh
set -e
chmod +x ./*.r 2>/dev/null || true
./test.r
