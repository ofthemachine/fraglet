#!/bin/sh
set -e

# Locks in the single-output shorthand: --output <dest> (no "=", not itself
# a declared relpath) means "the fraglet's one declared output, saved as
# <dest>" -- no need to repeat the fraglet's own internal filename back to
# it when there's only one candidate to disambiguate.

cat > one_output.py <<'EOF'
#: output=result.txt
open("/output/result.txt", "w").write("hello\n")
print("script ran")
EOF

echo "=== Test 1: bare dest, no '=', not the declared name -- shorthand applies ==="
fragletc --image 100hellos/python:latest --output my-name.txt one_output.py 2>&1
echo "my-name.txt exists: $([ -f my-name.txt ] && echo yes || echo no)"
cat my-name.txt
rm -f my-name.txt

echo ""
echo "=== Test 2: the real declared name still works, unchanged ==="
fragletc --image 100hellos/python:latest --output result.txt one_output.py 2>&1
echo "result.txt exists: $([ -f result.txt ] && echo yes || echo no)"
rm -f result.txt

echo ""
echo "=== Test 3: explicit relpath=dest with a WRONG relpath is a real error, not silently reinterpreted ==="
fragletc --image 100hellos/python:latest --output wrong.txt=out.txt one_output.py 2>&1 || true

cat > two_outputs.py <<'EOF'
#: output=first.txt
#: output=second.txt
open("/output/first.txt", "w").write("one\n")
open("/output/second.txt", "w").write("two\n")
print("script ran")
EOF

echo ""
echo "=== Test 4: two declared outputs -- ambiguous, shorthand does NOT apply ==="
fragletc --image 100hellos/python:latest --output my-name.txt two_outputs.py 2>&1 || true

rm -f one_output.py two_outputs.py
