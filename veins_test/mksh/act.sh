#!/bin/sh
set -e
chmod +x ./*.mksh 2>/dev/null || true
./test.mksh
