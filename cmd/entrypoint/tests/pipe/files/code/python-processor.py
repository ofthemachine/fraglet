#!/usr/bin/env -S fragletc --vein=python
import sys

# Python: Read lines, uppercase them, add "PYTHON:" prefix
for line in sys.stdin:
    print(f"PYTHON: {line.strip().upper()}")
