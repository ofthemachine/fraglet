#!/bin/sh
set -e
chmod +x ./*.wren 2>/dev/null || true
./test.wren
