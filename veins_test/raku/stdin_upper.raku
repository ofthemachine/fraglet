#!/usr/bin/env -S fragletc --vein=raku
for $*IN.lines() -> $line {
    say $line.uc;
}
