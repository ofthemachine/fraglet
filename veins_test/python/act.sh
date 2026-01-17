#!/bin/sh
set -e
chmod +x ./*.py 2>/dev/null || true
./test.py
