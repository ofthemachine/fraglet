#!/bin/sh
set -e

# Regression: relative output paths -- --output-dir=. baked into a shebang,
# and --output=<dest> -- must resolve against the caller's cwd at
# invocation, never the directory the script itself happens to live in.
# This falls out of exec(2) never chdir'ing (fragletc never calls os.Chdir
# either), so a script at skills/fun/meme.sh invoked as ./skills/fun/meme.sh
# from a project root writes relative output where the caller stands, the
# same way a locally installed CLI tool would -- its own location on disk
# is irrelevant to where a relative -o path lands.

mkdir -p skills/fun

cat > skills/fun/dir-mode.py <<'EOF'
#!/usr/bin/env -S fragletc --image=100hellos/python:latest --output-dir=.
open("/output/out.txt", "w").write("hello\n")
print("wrote out.txt")
EOF
chmod +x skills/fun/dir-mode.py

cat > skills/fun/declared-mode.py <<'EOF'
#!/usr/bin/env -S fragletc --image=100hellos/python:latest
#: output=result.txt
open("/output/result.txt", "w").write("declared\n")
print("wrote result.txt")
EOF
chmod +x skills/fun/declared-mode.py

echo "=== Test 1: --output-dir=. baked into the shebang resolves against the caller's cwd, not the script's directory ==="
./skills/fun/dir-mode.py
echo "landed in caller cwd: $([ -f out.txt ] && echo yes || echo no)"
echo "did not land in script dir: $([ -f skills/fun/out.txt ] && echo WRONG || echo correct)"

echo ""
echo "=== Test 2: --output=<dest> resolves against the caller's cwd, not the script's directory ==="
./skills/fun/declared-mode.py --output=image.png
echo "landed in caller cwd: $([ -f image.png ] && echo yes || echo no)"
echo "did not land in script dir: $([ -f skills/fun/image.png ] && echo WRONG || echo correct)"
