#!/usr/bin/env -S fragletc --vein=fortran
  character(256) :: line
  integer :: io
  read(*, '(a)', iostat=io) line
  if (io == 0) then
      call upper_case(line)
      write(*, '(a)') trim(line)
  end if
contains
  subroutine upper_case(s)
      character(*), intent(inout) :: s
      integer :: i
      do i = 1, len_trim(s)
          if (s(i:i) >= 'a' .and. s(i:i) <= 'z') s(i:i) = achar(iachar(s(i:i)) - 32)
      end do
  end subroutine
