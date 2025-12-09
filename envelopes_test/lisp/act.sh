#!/bin/sh
set -e

# Test Lisp envelope by name
FRAGLETC="./fragletc"

echo "=== Test: List processing ==="
cat <<'EOF' | "$FRAGLETC" --envelope lisp
(let ((numbers '(1 2 3 4 5)))
  (let ((squared (mapcar (lambda (x) (* x x)) numbers)))
    (format t "Sum of squares: ~a~%" (reduce #'+ squared))))
EOF

