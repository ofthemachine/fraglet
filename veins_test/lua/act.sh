#!/bin/sh
set -e
chmod +x ./*.lua 2>/dev/null || true
./test.lua
