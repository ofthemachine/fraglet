#!/bin/sh
set -e
chmod +x ./*.hs 2>/dev/null || true
./test.hs
