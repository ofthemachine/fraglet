#!/bin/sh
set -e
chmod +x ./*.fs 2>/dev/null || true
./test.fs
