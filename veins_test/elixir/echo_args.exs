#!/usr/bin/env -S fragletc --vein=elixir
args = System.argv()
IO.puts("Args: #{Enum.join(args, " ")}")
