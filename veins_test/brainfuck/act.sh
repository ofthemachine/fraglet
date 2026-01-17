#!/bin/sh
set -e
chmod +x ./*.bf 2>/dev/null || true
./test.bf
