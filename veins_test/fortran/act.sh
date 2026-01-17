#!/bin/sh
set -e
chmod +x ./*.f 2>/dev/null || true
./test.f
