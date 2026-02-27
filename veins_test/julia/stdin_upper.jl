#!/usr/bin/env -S fragletc --vein=julia
for line in eachline(stdin)
    println(uppercase(line))
end
