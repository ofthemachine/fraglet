#!/usr/bin/env -S fragletc --vein=scheme
(import (chibi))
(import (scheme base))
(import (chibi string))
(let loop ((line (read-line)))
  (when (not (eof-object? line))
    (display (string-upcase-ascii line))
    (newline)
    (loop (read-line))))
