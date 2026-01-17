#!/bin/sh
set -e
chmod +x ./*.rkt 2>/dev/null || true
./test.rkt
