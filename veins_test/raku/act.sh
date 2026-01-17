#!/bin/sh
set -e
chmod +x ./*.raku 2>/dev/null || true
./test.raku
