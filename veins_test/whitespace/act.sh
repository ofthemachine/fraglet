#!/bin/sh
set -e
chmod +x ./*.ws 2>/dev/null || true
./test.ws
