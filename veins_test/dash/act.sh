#!/bin/sh
set -e
chmod +x ./*.dash 2>/dev/null || true
./test.dash
