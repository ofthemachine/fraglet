#!/usr/bin/env -S fragletc --vein=fortran
  integer :: n, i
  character(256) :: arg
  character(1024) :: result
  n = command_argument_count()
  result = 'Args:'
  do i = 1, n
      call get_command_argument(i, arg)
      result = trim(result) // ' ' // trim(arg)
  end do
  print '(a)', trim(result)
