while IFS= read -r line; do echo "$line" | tr '[:lower:]' '[:upper:]'; done
