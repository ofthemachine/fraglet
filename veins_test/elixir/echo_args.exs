#!/usr/bin/env -S fragletc --vein=elixir
args = System.argv()
puts("Args: #{Enum.join(args, " ")}")
