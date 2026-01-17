#!/bin/sh
set -e
chmod +x ./*.nix 2>/dev/null || true
./test.nix
