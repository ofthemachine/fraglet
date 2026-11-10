#!/bin/sh
set -e

# A script that says how it runs (#: run=) needs no fraglet entrypoint in its
# image: the body lands where the line names it, params arrive under their
# bare names, and a malformed line is refused before anything runs.

echo "=== Test 1: the body runs where the line names it ==="
cat > hello.sh <<'SCRIPT'
#!/usr/bin/env -S fragletc --image alpine:latest
#: run=sh {code:/work/hello.sh}
#: param=name:default=nobody
echo "hello $NAME from $0"
SCRIPT
fragletc hello.sh -p name=ottoline

echo ""
echo "=== Test 2: a declared default arrives too ==="
fragletc hello.sh

echo ""
echo "=== Test 3: a run line with nowhere for the body is refused ==="
cat > bad.sh <<'SCRIPT'
#!/usr/bin/env -S fragletc --image alpine:latest
#: run=sh -s
echo never
SCRIPT
fragletc bad.sh 2>&1 || echo "exit: $?"

echo ""
echo "=== Test 4: lint says so ==="
fragletc lint bad.sh 2>&1 | grep run-value || true
