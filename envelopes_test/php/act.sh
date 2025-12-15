#!/bin/sh
set -e

FRAGLETC="./fragletc"

cat <<'EOF' | "$FRAGLETC" --envelope php
function isPrime(int $n): bool {
    if ($n < 2) return false;
    if ($n % 2 === 0) return $n === 2;
    for ($i = 3; $i * $i <= $n; $i += 2) {
        if ($n % $i === 0) return false;
    }
    return true;
}

$nums = range(10, 25);
$primes = array_values(array_filter($nums, 'isPrime'));
echo implode(",", $primes) . PHP_EOL;
EOF

