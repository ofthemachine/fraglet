#!/usr/bin/env -S fragletc --vein=crystal
STDIN.each_line do |line|
  puts line.upcase
end
