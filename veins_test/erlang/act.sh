#!/bin/sh
set -e
chmod +x ./*.erl 2>/dev/null || true
./test.erl
