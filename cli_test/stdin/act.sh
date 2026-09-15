#!/bin/sh
set -e

# Locks in the stdin contract: a piped stdin is buffered by default (read
# to EOF, hashed into the receipt, forwarded), #: stdin=stream / --stdin
# are the explicit alternatives, and a receipt's memo key exists only for
# a run whose inputs were all declared and recorded. An earlier draft shipped a
# silent "undeclared stdin is not forwarded" default because nothing here
# covered stdin at all.

cat > upper.py <<'PY'
#: d=uppercases stdin
import sys
print(sys.stdin.read().upper(), end="")
PY

cat > stream.py <<'PY'
#: d=uppercases stdin, declared as a live stream
#: stdin=stream
import sys
print(sys.stdin.read().upper(), end="")
PY

cat > args.py <<'PY'
#: d=echoes positional args
import sys
print(sys.argv[1:])
PY

echo "=== Test 1: undeclared stdin is buffered and forwarded ==="
echo "hello" | fragletc --image 100hellos/python:latest upper.py

echo ""
echo "=== Test 2: --stdin=none attaches nothing even when piped ==="
echo "hello" | fragletc --image 100hellos/python:latest upper.py --stdin=none
echo "(end)"

echo ""
echo "=== Test 3: a buffered run's receipt records the stdin hash and has a memo key ==="
echo "hello" | fragletc --image 100hellos/python:latest --receipt buffered.json upper.py 2>/dev/null
grep -c '"memo_key"' buffered.json
grep -o '"stdin_mode": "[a-z]*"' buffered.json
grep -o '"": "sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"' buffered.json

echo ""
echo "=== Test 4: a stream run is receipted honestly: no stdin hash, no memo key ==="
echo "hello" | fragletc --image 100hellos/python:latest --receipt stream.json stream.py 2>/dev/null
grep -c '"memo_key"' stream.json || true
grep -o '"stdin_mode": "[a-z]*"' stream.json
grep -o '"inputs": {}' stream.json

echo ""
echo "=== Test 5: positional args are recorded and cost the run its memo key ==="
fragletc --image 100hellos/python:latest --receipt args.json args.py foo 2>/dev/null
grep -c '"memo_key"' args.json || true
grep -A 2 '"argv"' args.json | tr -d ' \n'
echo ""

echo ""
echo "=== Test 6: --stdin rejects anything but none, buffer, stream ==="
fragletc --image 100hellos/python:latest upper.py --stdin=pipe 2>&1 || true

rm -f upper.py stream.py args.py buffered.json stream.json args.json
