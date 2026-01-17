#!/bin/sh
set -e
chmod +x ./*.ts 2>/dev/null || true
./test.ts
