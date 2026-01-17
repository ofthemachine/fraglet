#!/bin/sh
set -e
chmod +x ./*.vala 2>/dev/null || true
./test.vala
