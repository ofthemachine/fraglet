#!/bin/sh
set -e
chmod +x ./*.rs 2>/dev/null || true
./test.rs
