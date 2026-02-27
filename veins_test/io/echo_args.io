#!/usr/bin/env -S fragletc --vein=io
write("Args:")
System args rest foreach(arg, write(" " .. arg))
writeln
