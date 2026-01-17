#!/bin/sh
set -e
chmod +x ./*.io 2>/dev/null || true
./test.io
