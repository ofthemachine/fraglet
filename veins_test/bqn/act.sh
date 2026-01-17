#!/bin/sh
set -e
chmod +x ./*.bqn 2>/dev/null || true
./test.bqn
