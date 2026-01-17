#!/bin/sh
set -e
chmod +x ./*.bal 2>/dev/null || true
./test.bal
