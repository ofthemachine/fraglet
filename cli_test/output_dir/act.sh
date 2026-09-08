#!/bin/sh
set -e

# Locks in --output-dir behavior: everything written to /output is copied
# into the given host directory after the run, with no output= declaration
# needed at all — for wrapping a real CLI tool whose output filename is
# only known at runtime (an explicit flag it's passed, or its own default
# naming), unlike --output which requires the exact relpath declared ahead
# of time.

cat > dynamic_name.py <<'EOF'
import sys
name = sys.argv[1] if len(sys.argv) > 1 else "default.txt"
open(f"/output/{name}", "w").write("hi\n")
print(f"wrote {name}")
EOF

echo "=== Test 1: filename chosen at runtime, no output= declaration needed ==="
mkdir -p outdir
fragletc --image 100hellos/python:latest --output-dir=outdir dynamic_name.py custom-name.txt 2>&1
echo "custom-name.txt exists: $([ -f outdir/custom-name.txt ] && echo yes || echo no)"
cat outdir/custom-name.txt
rm -rf outdir

echo ""
echo "=== Test 2: no arg at all -- the script's own default name still lands ==="
mkdir -p outdir2
fragletc --image 100hellos/python:latest --output-dir=outdir2 dynamic_name.py 2>&1
echo "default.txt exists: $([ -f outdir2/default.txt ] && echo yes || echo no)"
rm -rf outdir2

echo ""
echo "=== Test 3: --output-dir and --output are mutually exclusive ==="
fragletc --image 100hellos/python:latest --output-dir=outdir3 --output default.txt dynamic_name.py 2>&1 || true
rm -rf outdir3

echo ""
echo "=== Test 4: same --output-dir, same filename, run twice -- overwrites, same as a locally installed tool would ==="
mkdir -p outdir4
fragletc --image 100hellos/python:latest --output-dir=outdir4 dynamic_name.py repeat.txt 2>&1
cat outdir4/repeat.txt
cat > dynamic_name_v2.py <<'EOF'
open("/output/repeat.txt", "w").write("second run\n")
print("wrote repeat.txt")
EOF
fragletc --image 100hellos/python:latest --output-dir=outdir4 dynamic_name_v2.py 2>&1
cat outdir4/repeat.txt
rm -rf outdir4 dynamic_name_v2.py

rm -f dynamic_name.py
