#!/usr/bin/env -S fragletc --vein=ruby
numbers = [1, 2, 3, 4, 5]
squared = numbers.map { |x| x**2 }
puts "Sum of squares: #{squared.sum}"
