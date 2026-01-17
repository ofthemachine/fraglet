#!/bin/sh
set -e
chmod +x ./*.ash 2>/dev/null || true
./test.ash
