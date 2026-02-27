#!/usr/bin/env -S fragletc --vein=lua
for line in io.lines() do
    print(string.upper(line))
end
