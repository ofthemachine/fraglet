#!/usr/bin/env -S fragletc --vein=wren
import "io" for Stdin

var line = Stdin.readLine()
if (line != null) {
  var upper = ""
  for (c in line) {
    var code = c.codePoints[0]
    if (code >= 97 && code <= 122) {
      upper = upper + String.fromCodePoint(code - 32)
    } else {
      upper = upper + c
    }
  }
  System.print(upper)
}
