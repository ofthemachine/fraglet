#!/usr/bin/env -S fragletc --vein=janet
(def line (file/read stdin :line))
(print (string/ascii-upper line))
