#!/bin/sh
set -e

# Test NASM x86_64 vein by name
FRAGLETC="./fragletc"

echo "=== Test: x86_64 assembly system calls ==="
cat <<'EOF' | "$FRAGLETC" --vein nasm-x86_64
          ; Write existing message to stdout
          mov rax, 1              ; sys_write
          mov rdi, 1              ; stdout
          mov rsi, message        ; address of string (from template)
          mov rdx, 13             ; length (13 bytes: "Hello World!" + newline)
          syscall
          
          ; Exit with code 0
          mov rax, 60
          xor rdi, rdi            ; exit code 0
          syscall
EOF
