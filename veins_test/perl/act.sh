#!/bin/sh
set -e
chmod +x ./*.pl 2>/dev/null || true
./test.pl
