#!/bin/sh
set -e

# Network mode tests verify container network isolation:
# - Default execution has eth0 (Docker default bridge)
# - --network none flag severs eth0
# - #: network=none header directive severs eth0 automatically
# - Explicit --network flag overrides the header directive
# - #: network=required keeps standard network

echo "=== Test 1: default execution has eth0 ==="
fragletc --vein python -c 'import os; print("eth0:", os.path.exists("/sys/class/net/eth0"))'

echo ""
echo "=== Test 2: explicit --network none severs eth0 ==="
fragletc --vein python --network none -c 'import os; print("eth0:", os.path.exists("/sys/class/net/eth0"))'

echo ""
echo "=== Test 3: #: network=none header automatically severs eth0 ==="
cat > hermetic.py <<'EOF'
#!/usr/bin/env -S fragletc
#: network=none
import os
print("eth0:", os.path.exists("/sys/class/net/eth0"))
EOF
fragletc --vein python hermetic.py

echo ""
echo "=== Test 4: explicit --network bridge overrides #: network=none ==="
fragletc --vein python --network bridge hermetic.py
rm -f hermetic.py

echo ""
echo "=== Test 5: #: network=required keeps eth0 ==="
cat > required.py <<'EOF'
#!/usr/bin/env -S fragletc
#: network=required
import os
print("eth0:", os.path.exists("/sys/class/net/eth0"))
EOF
fragletc --vein python required.py
rm -f required.py
