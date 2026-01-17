#!/bin/sh
set -e
chmod +x ./*.groovy 2>/dev/null || true
./test.groovy
