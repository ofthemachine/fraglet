#!/usr/bin/env -S fragletc --vein=python
import sys
for line in sys.stdin:
    print(line.strip().upper())
