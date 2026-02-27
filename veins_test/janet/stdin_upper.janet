#!/usr/bin/env -S fragletc --vein=janet
(def line (string/trim (file/read stdin :line)))
(print (string/ascii-upper line))
