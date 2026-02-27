#!/usr/bin/env -S fragletc --vein=elixir
input = IO.read(:stdio, :all)
IO.puts(String.upcase(String.trim(input)))
