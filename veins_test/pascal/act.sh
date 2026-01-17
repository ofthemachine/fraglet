#!/bin/sh
set -e
chmod +x ./*.pas 2>/dev/null || true
./test.pas
