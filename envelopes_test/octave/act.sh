#!/bin/sh
set -e

# Test Octave envelope by name
FRAGLETC="./fragletc"

echo "=== Test: MATLAB-compatible matrix operations ==="
cat <<'EOF' | "$FRAGLETC" --envelope octave
A = [1, 2, 3; 4, 5, 6];
B = [7, 8; 9, 10; 11, 12];
C = A * B;
printf("Matrix A:\n");
disp(A);
printf("Matrix product A*B:\n");
disp(C);
EOF


