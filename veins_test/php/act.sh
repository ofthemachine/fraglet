#!/bin/sh
set -e
chmod +x ./*.php 2>/dev/null || true
./test.php
