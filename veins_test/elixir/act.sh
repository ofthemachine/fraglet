#!/bin/sh
set -e

FRAGLETC="./fragletc"

cat <<'EOF' | "$FRAGLETC" --vein elixir
defmodule Math do
  def fact(n), do: Enum.reduce(1..n, 1, &*/2)
end

IO.puts(Enum.join(Enum.map(1..5, &Math.fact/1), ","))
EOF

