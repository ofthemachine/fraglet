#!/bin/sh
set -e
chmod +x ./*.v 2>/dev/null || true
./test.v
