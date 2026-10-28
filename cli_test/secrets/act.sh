#!/bin/sh
set -e

# #: secret=NAME -- the value comes from the caller's env var NAME and
# reaches the program only as the read-only file $NAME_FILE
# (/run/fraglet/secrets/NAME). It is never a container env var, never on
# the docker command line, and never in a receipt.

CODE='#: secret=FT_TOKEN:d=Test credential
#: secret-output=uncensored
import os
path = os.environ["FT_TOKEN_FILE"]
value = open(path).read()
print("path:", path)
print("value:", value)
print("as env var:", "FT_TOKEN" in os.environ)
print("value anywhere in environ:", any(value in v for v in os.environ.values()))'

echo "=== Test 1: missing secret fails host-side (exit 2) ==="
unset FT_TOKEN
fragletc --vein python -c "$CODE" 2>&1 || echo "exit=$?"

echo ""
echo "=== Test 2: delivered as a file, never as an env var (multi-line ok) ==="
export FT_TOKEN='line-one
line-two'
fragletc --vein python -c "$CODE"

echo ""
echo "=== Test 3: forwarding a declared secret with -e is refused ==="
fragletc --vein python -e FT_TOKEN -c "$CODE" 2>&1 || echo "exit=$?"

echo ""
echo "=== Test 4: the receipt never carries the value, and the run keeps its memo key ==="
export FT_TOKEN='receipt-sentinel-value'
fragletc --vein python --receipt secret-receipt.json -c "$CODE" >/dev/null 2>&1
grep -c receipt-sentinel-value secret-receipt.json || true
grep -c '"memo_key"' secret-receipt.json

echo ""
echo "=== Test 5: a secret in what the run writes: fail (the default), redact, uncensored ==="
# The script answers for what it prints: its header says what fragletc does
# when a secret's value appears in stdout, stderr or an output file, and the
# receipt says where it appeared, by name, in every mode. --streams saves
# stdout and stderr exactly as the receipt hashes them.
export FT_TOKEN='leak-sentinel-0123'
LEAK='#: secret=FT_TOKEN:d=Test credential
#: output=out.txt
import os, sys
v = open(os.environ["FT_TOKEN_FILE"]).read()
print("stdout says", v)
print("stderr says", v, file=sys.stderr)
open("/output/out.txt", "w").write("file says " + v)'
for mode in "" redact uncensored; do
  code="$LEAK"
  [ -n "$mode" ] && code="#: secret-output=$mode
$LEAK"
  rm -rf o && mkdir o
  echo "--- mode ${mode:-undeclared}"
  rm -rf st
  fragletc --vein python --output-dir o --streams st --receipt r.json -c "$code" >so.txt 2>err.txt </dev/null && echo "exit=0" || echo "exit=$?"
  sed 's/^/stdout: /' so.txt
  sed 's/^/stderr: /' err.txt | grep -v '^stderr: $'
  if [ -f o/out.txt ]; then sed 's/^/file: /' o/out.txt; echo; else echo "file: none (a failed run delivers no outputs)"; fi
  python3 -c '
import json
r = json.load(open("r.json"))
so = r.get("secret_output", {})
print("receipt: exit", r["exit_code"], "mode", so.get("mode"), "seen", [s["secret"] + " in " + s["in"] for s in so.get("seen", [])])
import hashlib, os
h = lambda p: "sha256:" + hashlib.sha256(open(p, "rb").read()).hexdigest()
print("receipt hashes what was delivered: stdout", r["stdout"]["hash"] == h("so.txt") == h("st/stdout"), "stderr", r["stderr"]["hash"] == h("st/stderr"), "file", (r["outputs"].get("out.txt") == h("o/out.txt")) if os.path.exists("o/out.txt") else "-")'
done
