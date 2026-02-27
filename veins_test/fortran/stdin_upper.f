#!/usr/bin/env -S fragletc --vein=fortran
  character(256) :: line
  integer :: io, ci
  read(*, '(a)', iostat=io) line
  do while (io == 0)
    do ci = 1, len_trim(line)
      if (line(ci:ci) >= 'a' .and. line(ci:ci) <= 'z') line(ci:ci) = achar(iachar(line(ci:ci)) - 32)
    end do
    print *, trim(line)
    read(*, '(a)', iostat=io) line
  end do
