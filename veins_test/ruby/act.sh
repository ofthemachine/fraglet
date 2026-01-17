#!/bin/sh
set -e
chmod +x ./*.rb 2>/dev/null || true
./test.rb
