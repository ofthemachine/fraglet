#!/bin/sh
set -e

# Test Lisp vein by name
FRAGLETC="./fragletc"

echo "=== Test: List processing ==="
cat <<'EOF' | "$FRAGLETC" --vein lisp
(let ((numbers '(1 2 3 4 5)))
  (let ((squared (mapcar (lambda (x) (* x x)) numbers)))
    (format t "Sum of squares: ~a~%" (reduce #'+ squared))))
EOF


