#!/usr/bin/env -S fragletc --vein=racket
(for ([line (in-lines)])
  (displayln (string-upcase line)))
