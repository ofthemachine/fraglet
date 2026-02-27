#!/bin/sh
set -e
chmod +x ./*.bash 2>/dev/null || true
./test.bash
