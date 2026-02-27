#!/usr/bin/env -S fragletc --vein=clojure
(println "Args:" (clojure.string/join " " *command-line-args*))
