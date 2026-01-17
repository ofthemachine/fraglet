#!/bin/sh
set -e
chmod +x ./*.janet 2>/dev/null || true
./test.janet
