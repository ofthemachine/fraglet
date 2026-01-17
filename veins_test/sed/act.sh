#!/bin/sh
set -e
chmod +x ./*.sed 2>/dev/null || true
./test.sed
