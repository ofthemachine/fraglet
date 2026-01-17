#!/bin/sh
set -e
chmod +x ./*.kt 2>/dev/null || true
./test.kt
