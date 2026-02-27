#!/usr/bin/env -S fragletc --vein=crystal
class Greeting
  def greet(@name : String)
    puts "Hello #{@name}!"
  end
end

g = Greeting.new()
g.greet("World")
