#!/usr/bin/env -S fragletc --vein=tcl
while {[gets stdin line] >= 0} { puts [string toupper $line] }
