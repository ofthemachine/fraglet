#!/bin/sh
set -e
chmod +x ./*.wat 2>/dev/null || true
./test.wat
