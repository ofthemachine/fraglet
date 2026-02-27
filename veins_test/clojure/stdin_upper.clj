#!/usr/bin/env -S fragletc --vein=clojure
(require '[clojure.string :as str])
(doseq [line (line-seq (java.io.BufferedReader. *in*))]
  (println (str/upper-case line)))
