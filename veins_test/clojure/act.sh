#!/bin/sh
set -e
chmod +x ./*.clj 2>/dev/null || true
./test.clj
