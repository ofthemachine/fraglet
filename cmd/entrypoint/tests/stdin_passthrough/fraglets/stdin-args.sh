#!/bin/bash
echo "args: $@"
while IFS= read -r line; do
  echo "stdin: $line"
done
