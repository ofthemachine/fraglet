#!/bin/sh
set -e
chmod +x ./*.go ./*.goz 2>/dev/null || true
./test.goz
