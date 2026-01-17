#!/bin/sh
set -e
chmod +x ./*.c 2>/dev/null || true
./test.c
