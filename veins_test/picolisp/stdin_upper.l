#!/usr/bin/env -S fragletc --vein=picolisp
(in NIL
   (until (eof)
      (prinl (uppc (line T)))))
