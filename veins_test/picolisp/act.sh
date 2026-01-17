#!/bin/sh
set -e
chmod +x ./*.l 2>/dev/null || true
./test.l
