#!/bin/sh
set -e
chmod +x ./*.scm 2>/dev/null || true
./test.scm
