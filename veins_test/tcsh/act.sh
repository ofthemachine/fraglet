#!/bin/sh
set -e
chmod +x ./*.tcsh 2>/dev/null || true
./test.tcsh
