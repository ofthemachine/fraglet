#!/bin/sh
set -e
chmod +x ./*.java 2>/dev/null || true
./main.java f bar "fragletc is wonderful-see?"
./wordset.java
