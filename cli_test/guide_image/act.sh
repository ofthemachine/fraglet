#!/bin/sh
set -e

docker build --platform linux/amd64 -t fraglet-guide-image-cli:local -f Dockerfile . >/dev/null 2>&1

IMG=fraglet-guide-image-cli:local

echo "=== guide: --image flag (no vein) ==="
unset FRAGLET_VEINS_PATH
fragletc guide --image="$IMG"

echo ""
echo "=== guide: --mode and --image any order ==="
fragletc guide --mode=testmode --image="$IMG"
fragletc guide --image="$IMG" --mode=testmode


echo ""
echo "=== guide: -i short form ==="
unset FRAGLET_VEINS_PATH
fragletc guide -i "$IMG"

echo ""
echo "=== guide: --mode after -i ==="
fragletc guide -i "$IMG" --mode=testmode

echo ""
echo "=== essence: -i only ==="
unset FRAGLET_VEINS_PATH
fragletc essence -i "$IMG"

echo ""
echo "=== essence: --mode and --image any order ==="
fragletc essence --mode=testmode --image="$IMG"
fragletc essence --image="$IMG" --mode=testmode

echo ""
echo "=== vein-only still works ==="
export FRAGLET_VEINS_PATH=./veins.yml
fragletc guide guidetester

echo ""
echo "=== XOR: vein + --image must fail ==="
set +e
out=$(fragletc guide --image="$IMG" guidetester 2>&1)
code=$?
set -e
echo "$out"
if [ "$code" -eq 0 ]; then
  echo "expected nonzero exit for XOR" >&2
  exit 1
fi

docker rmi fraglet-guide-image-cli:local >/dev/null 2>&1 || true
