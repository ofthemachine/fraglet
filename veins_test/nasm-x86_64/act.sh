#!/bin/sh
set -e
chmod +x ./*.asm 2>/dev/null || true
./test.asm
