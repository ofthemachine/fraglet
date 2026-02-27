#!/usr/bin/env -S fragletc --vein=wat
  ;; String data: "Hello World!"
  (data (i32.const 0) "Hello World!\n")

  ;; WASI entry point
  (func $main (export "_start")
    ;; Set up iovec for fd_write
    ;; iovec.buf = 0 (pointer to string)
    (i32.store (i32.const 16) (i32.const 0))
    ;; iovec.len = 13 (length of "Hello World!\n")
    (i32.store (i32.const 20) (i32.const 13))

    ;; Call fd_write(1, 16, 1, 24)
    ;; 1 = stdout, 16 = pointer to iovec, 1 = number of iovecs, 24 = pointer to store result
    (call $fd_write (i32.const 1) (i32.const 16) (i32.const 1) (i32.const 24))
    drop

    ;; Exit with status 0
    (call $proc_exit (i32.const 0))
  )
