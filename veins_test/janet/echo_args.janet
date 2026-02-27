#!/usr/bin/env -S fragletc --vein=janet
(def all-args (dyn :args))
(def args (tuple/slice all-args 1))
(print "Args: " (string/join args " "))
