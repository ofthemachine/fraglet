#!/usr/bin/env -S fragletc --vein=elixir
input = IO.read(:stdio, :all)
puts(String.upcase(String.trim(input)))
