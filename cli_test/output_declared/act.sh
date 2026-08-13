#!/bin/sh
set -e

# Locks in the declared-output= behavior: /output is mounted whenever the
# fraglet declares any output=, independent of --output; every successfully
# copied file is confirmed by name (declared relpath AND actual host
# destination — a rename must never be silently invisible); every declared
# output not requested is reported as discarded, whether zero or only some
# were requested; no discard note at all when every declared output is
# requested.

cat > declare_two.py <<'EOF'
#: d=writes two declared outputs
#: output=first.txt
#: output=second.txt

open("/output/first.txt", "w").write("first\n")
open("/output/second.txt", "w").write("second\n")
print("script ran")
EOF

echo "=== Test 1: no --output at all (must run, not crash; both discarded) ==="
fragletc --image 100hellos/python:latest declare_two.py 2>&1
echo "first.txt exists: $([ -f first.txt ] && echo yes || echo no)"
echo "second.txt exists: $([ -f second.txt ] && echo yes || echo no)"

echo ""
echo "=== Test 2: --output second.txt only (first.txt must be reported, not silently dropped) ==="
fragletc --image 100hellos/python:latest --output second.txt declare_two.py 2>&1
echo "first.txt exists: $([ -f first.txt ] && echo yes || echo no)"
echo "second.txt exists: $([ -f second.txt ] && echo yes || echo no)"
cat second.txt
rm -f second.txt

echo ""
echo "=== Test 3: --output for both (no discard note at all) ==="
fragletc --image 100hellos/python:latest --output first.txt --output second.txt declare_two.py 2>&1
echo "first.txt exists: $([ -f first.txt ] && echo yes || echo no)"
echo "second.txt exists: $([ -f second.txt ] && echo yes || echo no)"
cat first.txt
cat second.txt
rm -f first.txt second.txt

echo ""
echo "=== Test 4: --output with a rename (=hostdest) must be confirmed under the ACTUAL name, not the declared one ==="
fragletc --image 100hellos/python:latest --output second.txt=renamed.txt declare_two.py 2>&1
echo "second.txt exists: $([ -f second.txt ] && echo yes || echo no)"
echo "renamed.txt exists: $([ -f renamed.txt ] && echo yes || echo no)"
cat renamed.txt
rm -f renamed.txt

rm -f declare_two.py
