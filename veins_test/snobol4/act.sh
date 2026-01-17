#!/bin/sh
set -e
chmod +x ./*.sno 2>/dev/null || true
./test.sno
