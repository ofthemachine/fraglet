#!/bin/sh
set -e

ENTRYPOINT="$(pwd)/fraglet-entrypoint"

echo "=== Python: stdin pipe ==="
echo "hello from python" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-py.py:/FRAGLET:ro" \
  100hellos/python:latest 2>&1

echo ""
echo "=== Python: multi-line stdin ==="
printf "alpha\nbeta\ngamma\n" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-py.py:/FRAGLET:ro" \
  100hellos/python:latest 2>&1

echo ""
echo "=== Ruby: stdin pipe ==="
echo "hello from ruby" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-rb.rb:/FRAGLET:ro" \
  100hellos/ruby:latest 2>&1

echo ""
echo "=== C: stdin pipe ==="
echo "hello from c" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-c.c:/FRAGLET:ro" \
  100hellos/the-c-programming-language:latest 2>&1

echo ""
echo "=== C++: stdin pipe ==="
echo "hello from cpp" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-cpp.cpp:/FRAGLET:ro" \
  100hellos/cpp:latest 2>&1

echo ""
echo "=== Java: stdin pipe ==="
echo "hello from java" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-java.java:/FRAGLET:ro" \
  100hellos/java:latest 2>&1

echo ""
echo "=== C#: stdin pipe ==="
echo "hello from csharp" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-cs.cs:/FRAGLET:ro" \
  -e DOTNET_NOLOGO=1 \
  -e DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 \
  100hellos/csharp:latest 2>/dev/null

echo ""
echo "=== Python: stdin + args ==="
echo "piped-data" | docker run --rm -i \
  -v "$ENTRYPOINT:/fraglet-entrypoint:ro" \
  -v "$(pwd)/fraglets/stdin-args-py.py:/FRAGLET:ro" \
  100hellos/python:latest arg1 arg2 2>&1
