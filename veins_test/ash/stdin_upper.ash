#!/usr/bin/env -S fragletc --vein=ash
while read -r line; do echo "$line" | tr "a-z" "A-Z"; done
