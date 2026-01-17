#!/bin/sh
set -e
chmod +x ./*.ha 2>/dev/null || true
./test.ha
