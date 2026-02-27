#!/usr/bin/env -S fragletc --vein=lisp
(loop for line = (read-line *standard-input* nil)
      while line
      do (format t "~A~%" (string-upcase line)))
