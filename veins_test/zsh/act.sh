#!/bin/sh
set -e
chmod +x ./*.zsh 2>/dev/null || true
./test.zsh
