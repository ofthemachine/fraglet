#!/bin/sh
set -e

# Binary is built for linux/amd64; run it inside Docker so it works on any host (darwin, etc.).
# Use absolute paths in config so cwd doesn't matter (avoid "no such file" in container).
sed 's|code/hello-world.sh|/work/code/hello-world.sh|g; s|^fragletTempPath: FRAGLET|fragletTempPath: /work/FRAGLET|; s|^guide: guide.md|guide: /work/guide.md|' fraglet.yaml > fraglet-docker.yaml

docker run --rm --platform linux/amd64 \
  -v "$(pwd):/work" -e FRAGLET_CONFIG=/work/fraglet-docker.yaml \
  alpine:latest /work/fraglet-entrypoint

echo "---"
docker run --rm --platform linux/amd64 \
  -v "$(pwd):/work" -e FRAGLET_CONFIG=/work/fraglet.yaml \
  alpine:latest /work/fraglet-entrypoint usage
echo "---"
docker run --rm --platform linux/amd64 \
  -v "$(pwd):/work" -e FRAGLET_CONFIG=/work/fraglet.yaml \
  alpine:latest /work/fraglet-entrypoint guide
