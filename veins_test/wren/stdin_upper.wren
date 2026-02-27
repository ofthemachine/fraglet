#!/usr/bin/env -S fragletc --vein=wren
import "io" for Stdin
var line = Stdin.readLine()
var upper = ""
for (c in line.codePoints) {
  if (c >= 97 && c <= 122) {
    upper = upper + String.fromCodePoint(c - 32)
  } else {
    upper = upper + String.fromCodePoint(c)
  }
}
System.print(upper)
