#!/bin/sh
set -e

# --params runs a script with a run's recorded params, from one JSON file,
# the script untouched; a name given with -p is taken from -p.

cat > greet.sh <<'SCRIPT'
#!/usr/bin/env -S fragletc --image alpine:latest
#: run=sh {code:/work/greet.sh}
#: param=name:default=nobody
#: param=times:default=1
#: param=note:file
echo "hello $NAME x$TIMES: $(cat "$NOTE")"
SCRIPT
echo "a note from a file" > note.txt
cat > greet.params.json <<'JSON'
{"name": "ottoline", "times": 3, "note": "note.txt"}
JSON

echo "=== Test 1: every param from the file ==="
fragletc --params greet.params.json greet.sh

echo ""
echo "=== Test 2: -p wins over the file ==="
fragletc --params=greet.params.json greet.sh -p name=gladys

echo ""
echo "=== Test 3: a file that is not an object is refused ==="
echo '[1, 2]' > bad.json
fragletc --params bad.json greet.sh 2>&1 || echo "exit: $?"
