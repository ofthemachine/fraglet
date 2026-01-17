#!/bin/sh
set -e
chmod +x ./*.fth 2>/dev/null || true
./test.fth
