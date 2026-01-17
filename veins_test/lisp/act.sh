#!/bin/sh
set -e
chmod +x ./*.lisp 2>/dev/null || true
./test.lisp
