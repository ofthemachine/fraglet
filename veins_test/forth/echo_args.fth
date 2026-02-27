#!/usr/bin/env -S fragletc --vein=forth
: main
  ." Args:"
  begin next-arg dup while
    space type
  repeat 2drop cr ;
main bye
