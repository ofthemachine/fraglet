#!/bin/sh
set -e
chmod +x ./*.unl 2>/dev/null || true
./test.unl
