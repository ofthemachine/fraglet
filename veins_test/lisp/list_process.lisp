#!/usr/bin/env -S fragletc --vein=lisp
(let ((numbers '(1 2 3 4 5)))
  (let ((squared (mapcar (lambda (x) (* x x)) numbers)))
    (format t "Sum of squares: ~a~%" (reduce #'+ squared))))
