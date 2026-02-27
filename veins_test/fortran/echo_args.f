#!/usr/bin/env -S fragletc --vein=fortran
  integer :: n, i
  character(256) :: arg
  n = iargc()
  write(*, '(a)', advance='no') 'Args:'
  do i = 1, n
    call getarg(i, arg)
    write(*, '(a)', advance='no') ' ' // trim(arg)
  end do
  print *
