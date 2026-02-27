#!/usr/bin/env -S fragletc --vein=nim
import std/strutils
try:
  while true:
    let line = readLine(stdin)
    echo line.toUpper()
except EOFError:
  discard
