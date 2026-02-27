#!/usr/bin/env -S fragletc --vein=forth
: upcase ( c -- c' )
  dup [char] a [char] z 1+ within if 32 - then ;
: read-and-upcase
  pad 256 stdin read-line throw drop
  pad swap bounds ?do
    i c@ upcase emit
  loop cr ;
read-and-upcase bye
