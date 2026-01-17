#!/bin/sh
set -e
chmod +x ./*.exs 2>/dev/null || true
./test.exs
