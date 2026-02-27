#!/usr/bin/env -S fragletc --vein=octave
line = fgetl(stdin);
printf("%s\n", toupper(line));
