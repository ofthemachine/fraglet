#!/bin/sh
set -e

ENTRYPOINT="$(pwd)/fraglet-entrypoint"

echo "=== Bash: stdin pipe ==="
echo "hello from bash" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-bash.bash:/FRAGLET:ro" \
  100hellos/bash:latest 2>&1

echo ""
echo "=== Go: stdin pipe ==="
echo "STDINWORKS" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-go.go.fraglet:/FRAGLET:ro" \
  100hellos/golang:latest 2>&1

echo ""
echo "=== Rust: stdin pipe ==="
echo "hello from rust" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-rust.rs:/FRAGLET:ro" \
  100hellos/rust:latest 2>&1

echo ""
echo "=== Haskell: stdin pipe ==="
echo "hello from haskell" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-hs.hs:/FRAGLET:ro" \
  100hellos/haskell:latest 2>&1

echo ""
echo "=== Perl: stdin pipe ==="
echo "hello from perl" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-perl.pl:/FRAGLET:ro" \
  100hellos/perl:latest 2>&1

echo ""
echo "=== Lua: stdin pipe ==="
echo "hello from lua" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-lua.lua:/FRAGLET:ro" \
  100hellos/lua:latest 2>&1

echo ""
echo "=== Kotlin: stdin pipe ==="
echo "hello from kotlin" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-kotlin.kt:/FRAGLET:ro" \
  100hellos/kotlin:latest 2>&1

echo ""
echo "=== Scala: stdin pipe ==="
echo "hello from scala" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-scala.scala:/FRAGLET:ro" \
  100hellos/scala:latest 2>&1
