#!/bin/sh
set -e

# Locks in param=<alias>:file (the input-side counterpart to
# output_declared/): a param declared with the :file modifier resolves its
# CLI value as a HOST path, mounts it read-only into the container at the
# fixed path /input/<alias>, and rewrites the transported value to that
# container path — the fraglet never sees the host path itself, only the
# fixed mount convention. A missing host file must fail before the
# container ever starts, and the mount must stay read-only.

cat > read_one.py <<'EOF'
#: d=reads a single file-shaped param
#: param=doc:file

with open("/input/doc") as f:
    print(f.read().strip())
EOF

cat > read_two.py <<'EOF'
#: d=reads two file-shaped params, independently mounted
#: param=doc:file
#: param=notes:file

with open("/input/doc") as f:
    print("doc:", f.read().strip())
with open("/input/notes") as f:
    print("notes:", f.read().strip())
EOF

cat > try_write.py <<'EOF'
#: d=confirms the file-shaped mount is read-only
#: param=doc:file

try:
    open("/input/doc", "a").write("nope")
    print("write succeeded (BUG: should be read-only)")
except OSError as e:
    print("write blocked:", "Read-only file system" in str(e))
EOF

echo "hello from host" > doc.txt
echo "second file" > notes.txt

echo "=== Test 1: single file-shaped param ==="
fragletc --image 100hellos/python:latest -p doc=doc.txt read_one.py

echo ""
echo "=== Test 2: two file-shaped params, independently mounted ==="
fragletc --image 100hellos/python:latest -p doc=doc.txt -p notes=notes.txt read_two.py

echo ""
echo "=== Test 3: missing host file fails before the container ever runs ==="
fragletc --image 100hellos/python:latest -p doc=does-not-exist.txt read_one.py 2>&1 || true

echo ""
echo "=== Test 4: mount is read-only ==="
fragletc --image 100hellos/python:latest -p doc=doc.txt try_write.py

rm -f doc.txt notes.txt read_one.py read_two.py try_write.py
