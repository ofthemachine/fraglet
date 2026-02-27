#!/usr/bin/env -S fragletc --vein=r-project
input <- readLines(con = "stdin")
cat(toupper(input), "\n")
