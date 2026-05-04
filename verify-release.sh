#!/bin/sh
set -e

REPO="ofthemachine/fraglet"

usage() {
  cat <<EOF
verify-release.sh — reproduce a release build in a pinned Go container

Usage:
  ./verify-release.sh <tag>

Examples:
  ./verify-release.sh fragletc-v0.6.0
  ./verify-release.sh entrypoint-v0.5.0
  ./verify-release.sh entrypoint-v0.6.0

Runs the build inside a golang Docker container matching the Go version
in go.mod at the tagged ref. Outputs sha256 checksums for all built
platform binaries — compare against the release's checksums.txt.
EOF
  exit 1
}

TAG="${1:-}"
[ -z "$TAG" ] && usage

case "$TAG" in
  fragletc-v*)    ARTIFACT="fragletc" ;;
  entrypoint-v*) ARTIFACT="entrypoint" ;;
  *) echo "error: unrecognized tag format: $TAG" >&2; exit 1 ;;
esac

git rev-parse "${TAG}^{commit}" >/dev/null 2>&1 \
  || { echo "error: tag $TAG not found. Run: git fetch --tags" >&2; exit 1; }

REPO_ROOT=$(git rev-parse --show-toplevel)

GO_VERSION=$(git show "${TAG}:go.mod" | grep "^go " | awk '{print $2}')
[ -z "$GO_VERSION" ] && { echo "error: could not read Go version from go.mod at ${TAG}" >&2; exit 1; }
GO_IMAGE="golang:${GO_VERSION}-bookworm"

echo "Using ${GO_IMAGE}" >&2

docker run --rm \
  -v "${REPO_ROOT}:/src:ro" \
  -e TAG="$TAG" \
  -e ARTIFACT="$ARTIFACT" \
  -e REPO="$REPO" \
  -w /build \
  "${GO_IMAGE}" sh -c '
    set -e

    git clone --quiet /src /build
    cd /build
    git checkout "$TAG" --quiet

    RELEASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

    if [ "$ARTIFACT" = "fragletc" ]; then
      # Detect layout at this ref
      if [ -d "cmd/fragletc" ]; then
        TARGET="./cmd/fragletc"
        INFO_PATH="cmd/fragletc/build-info.json"
      elif [ -d "cli" ]; then
        TARGET="./cli"
        INFO_PATH="cli/build-info.json"
      else
        echo "error: cannot find fragletc source" >&2; exit 1
      fi

      curl -fsSL -o "$INFO_PATH" "${RELEASE_URL}/build-info.json" \
        || { echo "error: could not download build-info.json" >&2; exit 1; }

      PLATFORMS="linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64"
      for platform in $PLATFORMS; do
        GOOS=$(echo "$platform" | cut -d/ -f1)
        GOARCH=$(echo "$platform" | cut -d/ -f2)
        OUT="fragletc-${GOOS}-${GOARCH}"
        [ "$GOOS" = "windows" ] && OUT="${OUT}.exe"

        CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
          -trimpath -buildvcs=false -ldflags="-s -w" \
          -o "$OUT" $TARGET
      done

      sha256sum fragletc-* | grep -v "\.sha256"

    elif [ "$ARTIFACT" = "entrypoint" ]; then
      if [ -d "cmd/entrypoint" ]; then
        TARGET="./cmd/entrypoint"
      elif [ -f "entrypoint/go.mod" ]; then
        cd entrypoint
        TARGET="./cmd"
      else
        echo "error: cannot find entrypoint source" >&2; exit 1
      fi

      PLATFORMS="linux/amd64 linux/arm64"
      for platform in $PLATFORMS; do
        GOOS=$(echo "$platform" | cut -d/ -f1)
        GOARCH=$(echo "$platform" | cut -d/ -f2)
        OUT="fraglet-entrypoint-${GOOS}-${GOARCH}"

        CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
          -trimpath -buildvcs=false -ldflags="-s -w" \
          -o "$OUT" $TARGET
      done

      sha256sum fraglet-entrypoint-*
    fi
  '
