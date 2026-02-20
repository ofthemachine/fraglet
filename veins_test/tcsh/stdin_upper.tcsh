#!/usr/bin/env -S fragletc --vein=tcsh
set line = $<
echo "$line" | tr "a-z" "A-Z"
