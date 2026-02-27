#!/usr/bin/env -S fragletc --vein=lisp
(format t "Args: ~{~a~^ ~}~%" (cdr sb-ext:*posix-argv*))
