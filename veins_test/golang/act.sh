#!/bin/sh
set -e
chmod +x ./*.go 2>/dev/null || true
./test.go
