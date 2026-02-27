#!/usr/bin/env -S fragletc --vein=scheme
(import (scheme base))
(import (scheme write))
(import (chibi string))
(let loop ((line (read-line)))
  (when (not (eof-object? line))
    (display (string-upcase-ascii line))
    (newline)
    (loop (read-line))))
