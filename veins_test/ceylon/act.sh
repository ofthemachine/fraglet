#!/bin/sh
set -e
chmod +x ./*.ceylon 2>/dev/null || true
./test.ceylon
