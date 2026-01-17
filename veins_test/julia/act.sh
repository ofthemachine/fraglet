#!/bin/sh
set -e
chmod +x ./*.jl 2>/dev/null || true
./test.jl
