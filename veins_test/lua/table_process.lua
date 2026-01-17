#!/usr/bin/env -S fragletc --vein=lua
local numbers = {1, 2, 3, 4, 5}
local sum = 0
for i, value in ipairs(numbers) do
  sum = sum + value * value
end
print(string.format("Sum of squares: %d", sum))
