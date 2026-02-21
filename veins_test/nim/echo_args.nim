#!/usr/bin/env -S fragletc --vein=nim
import std/strutils
import std/os
var a: seq[string]
for i in 1..paramCount(): a.add paramStr(i)
echo "Args: ", a.join(" ")
