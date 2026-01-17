#!/bin/sh
set -e
chmod +x ./*.emojic 2>/dev/null || true
./test.emojic
