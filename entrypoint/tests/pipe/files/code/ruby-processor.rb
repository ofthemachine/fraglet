#!/usr/bin/env -S fragletc --vein=ruby
# Ruby: Read lines, add line numbers and "RUBY:" prefix
$stdin.each_line.with_index(1) do |line, index|
  puts "RUBY[#{index}]: #{line.chomp}"
end
