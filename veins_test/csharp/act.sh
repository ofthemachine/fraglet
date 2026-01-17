#!/bin/sh
set -e
chmod +x ./*.cs 2>/dev/null || true
./test.cs
