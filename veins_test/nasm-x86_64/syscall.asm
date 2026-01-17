#!/usr/bin/env -S fragletc --vein=nasm-x86_64
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
