#!/bin/sh
set -e
chmod +x ./*.zig 2>/dev/null || true
./test.zig
