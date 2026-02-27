#!/usr/bin/env -S fragletc --vein=nasm-x86_64
mov       r12, [rsp]
mov       r13, rsp
add       r13, 8
mov       rdi, 1
lea       rsi, [rel args_prefix]
mov       rdx, 6
mov       rax, 1
syscall
cmp       r12, 2
jl        .done
mov       r14, 1
.loop:
mov       rsi, [r13 + r14*8]
xor       rdx, rdx
.strlen:
cmp       byte [rsi + rdx], 0
je        .print
inc       rdx
jmp       .strlen
.print:
mov       rdi, 1
mov       rax, 1
syscall
inc       r14
cmp       r14, r12
jge       .done
mov       rdi, 1
lea       rsi, [rel space_char]
mov       rdx, 1
mov       rax, 1
syscall
jmp       .loop
.done:
lea       rsi, [rel newline_char]
mov       rdx, 1
mov       rdi, 1
mov       rax, 1
syscall
mov       rax, 60
xor       rdi, rdi
syscall
section   .data
args_prefix: db "Args: "
space_char:  db " "
newline_char: db 10
