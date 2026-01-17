#!/bin/sh
set -e
chmod +x ./*.fnl 2>/dev/null || true
./test.fnl
