#!/bin/sh
set -e
chmod +x ./*.scala 2>/dev/null || true
./test.scala
