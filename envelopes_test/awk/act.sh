#!/bin/sh
set -e

FRAGLETC="./fragletc"

cat <<'EOF' | "$FRAGLETC" --envelope awk
BEGIN {
  data = "3 5 8 13"
  n = split(data, arr, " ")
  sum = 0
  for (i = 1; i <= n; i++) {
    sum += arr[i]
  }
  printf("sum=%d; mean=%.2f\n", sum, sum / n)
}
EOF

