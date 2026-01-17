#!/bin/sh
set -e
chmod +x ./*.sml 2>/dev/null || true
./test.sml
