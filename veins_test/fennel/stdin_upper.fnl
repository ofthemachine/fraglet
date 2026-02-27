#!/usr/bin/env -S fragletc --vein=fennel
(each [line (io.lines)]
  (print (string.upper line)))
