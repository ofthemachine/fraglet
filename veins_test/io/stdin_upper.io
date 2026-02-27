#!/usr/bin/env -S fragletc --vein=io
loop(
  line := File standardInput readLine
  if(line isNil, break)
  line asUppercase println
)
