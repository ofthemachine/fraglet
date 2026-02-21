#!/usr/bin/env -S fragletc --vein=racket
(require racket/string)
(display "Args: ")
(displayln (string-join (vector->list (current-command-line-arguments)) " "))
