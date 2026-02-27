#!/bin/sh
set -e
chmod +x ./*.arnoldc 2>/dev/null || true
./test.arnoldc
