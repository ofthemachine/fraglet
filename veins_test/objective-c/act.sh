#!/bin/sh
set -e
chmod +x ./*.m 2>/dev/null || true
./test.m
