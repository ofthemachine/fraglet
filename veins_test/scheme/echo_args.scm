#!/usr/bin/env -S fragletc --vein=scheme
(import (scheme base))
(import (scheme write))
(import (scheme process-context))
(import (chibi string))
(display "Args: ")
(display (string-join (cdr (command-line)) " "))
(newline)
