#!/usr/bin/env -S fragletc --vein=wat
;; Store string
(data (i32.const 0) "Hello fraglet!\n")

;; WASI entry point
(func $main (export "_start")
  ;; Set up iovec
  (i32.store (i32.const 16) (i32.const 0))
  (i32.store (i32.const 20) (i32.const 15))
  (call $fd_write (i32.const 1) (i32.const 16) (i32.const 1) (i32.const 24))
  drop

  ;; Exit with status 0
  (call $proc_exit (i32.const 0))
)
