#!/usr/bin/env -S fragletc --vein=janet
(def args (dyn :args))
(printf "Args: %s" (string/join (array/slice args 1) " "))
(print)
