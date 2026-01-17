#!/bin/sh
set -e
chmod +x ./*.tcl 2>/dev/null || true
./test.tcl
