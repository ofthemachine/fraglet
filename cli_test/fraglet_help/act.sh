#!/bin/sh
set -e
# Shebang is #!/usr/bin/env -S fragletc --vein=python — fragletc must be on PATH (clitest provides it).
chmod +x ./*.py 2>/dev/null || true

echo "=== Test: shebang fraglet ./params_example.py --fraglet-help ==="
./params_example.py --fraglet-help

echo ""
echo "=== Test: shebang fraglet ./multiline_meta.py --fraglet-help ==="
./multiline_meta.py --fraglet-help

echo ""
echo "=== Test: shebang fraglet absolute path .../multiline_meta.py --fraglet-help ==="
"$(pwd)/multiline_meta.py" --fraglet-help

echo ""
echo "=== Test: shebang fraglet ./bare.py --fraglet-help ==="
./bare.py --fraglet-help

echo ""
echo "=== Test: fragletc --fraglet-help -c with fraglet-meta ==="
fragletc --fraglet-help -c '# fraglet-meta: param=port:envvar=HURL_VARIABLE_port:default=8080'

echo ""
echo "=== Test: fragletc --fraglet-help -c without meta ==="
fragletc --fraglet-help -c 'print("hello")'

echo ""
echo "=== Test: fragletc --fraglet-help with no script or -c ==="
fragletc --fraglet-help 2>&1 || true

echo ""
echo "=== Test: shebang fraglet ./dup_params.py --fraglet-help ==="
./dup_params.py --fraglet-help

echo ""
echo "=== Test: shebang fraglet forwards script args (only --fraglet-help is special) ==="
./prints_argv.py --scheme https api.example

echo ""
echo "=== Test: -p after script path is gobbled (fraglet-meta transport, not argv) ==="
./prints_argv.py tail -p ghost=value

echo ""
echo "=== Test: program flags after script stay argv; -p stripped in the middle ==="
./prints_argv.py --profile prod -p city=paris extra

echo ""
echo "=== Test: after -- , -p passes through to program argv (not gobbled) ==="
./prints_argv.py -- -p ghost=value
