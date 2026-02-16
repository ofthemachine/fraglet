#!/usr/bin/env -S fragletc --vein=bash
while read -r line; do echo "$line" | tr "a-z" "A-Z"; done
