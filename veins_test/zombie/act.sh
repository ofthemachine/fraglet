#!/bin/sh
set -e
chmod +x ./*.zombie 2>/dev/null || true
./test.zombie
