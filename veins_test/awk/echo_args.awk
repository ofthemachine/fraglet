#!/usr/bin/env -S fragletc --vein=awk
BEGIN {
  for (i = 1; i < ARGC; i++)
    print ARGV[i]
  ARGC = 1
}
