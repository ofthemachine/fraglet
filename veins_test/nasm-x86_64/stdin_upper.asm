#!/usr/bin/env -S fragletc --vein=nasm-x86_64
sub       rsp, 64
mov       rax, 0
mov       rdi, 0
mov       rsi, rsp
mov       rdx, 64
syscall
mov       rdx, rax
mov       rax, 1
mov       rdi, 1
mov       rsi, rsp
syscall
add       rsp, 64
mov       rax, 60
xor       rdi, rdi
syscall
