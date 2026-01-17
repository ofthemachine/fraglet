#!/bin/sh
set -e
chmod +x ./*.js 2>/dev/null || true
./test.js
