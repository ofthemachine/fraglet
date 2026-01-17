#!/bin/sh
set -e
chmod +x ./*.dart 2>/dev/null || true
./test.dart
