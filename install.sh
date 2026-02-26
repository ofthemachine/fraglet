#!/bin/sh
set -e

REPO="ofthemachine/fraglet"
BINARY="fragletc"
INSTALL_DIR="${FRAGLETC_INSTALL_DIR:-$HOME/.local/bin}"

# --- Colors (disabled when not a terminal) ---
if [ -t 1 ]; then
  BOLD='\033[1m'
  DIM='\033[2m'
  GREEN='\033[0;32m'
  CYAN='\033[0;36m'
  YELLOW='\033[0;33m'
  RED='\033[0;31m'
  RESET='\033[0m'
else
  BOLD='' DIM='' GREEN='' CYAN='' YELLOW='' RED='' RESET=''
fi

info()  { printf "${CYAN}=>${RESET} %s\n" "$1"; }
ok()    { printf "${GREEN}=>${RESET} %s\n" "$1"; }
warn()  { printf "${YELLOW}warn:${RESET} %s\n" "$1"; }
fail()  { printf "${RED}error:${RESET} %s\n" "$1" >&2; exit 1; }

# --- Banner ---
printf "\n"
printf "${BOLD}  fragletc${RESET} ${DIM}— run code fragments in containers${RESET}\n"
printf "${DIM}  https://github.com/${REPO}${RESET}\n"
printf "\n"

# --- Platform detection ---
detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) fail "Unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64)  echo "arm64" ;;
    *) fail "Unsupported architecture: $(uname -m)" ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
PLATFORM="${OS}-${ARCH}"

info "Detected platform: ${BOLD}${PLATFORM}${RESET}"

# --- Find latest release ---
info "Finding latest release..."

RELEASES_URL="https://api.github.com/repos/${REPO}/releases/latest"

if command -v curl >/dev/null 2>&1; then
  RELEASE_JSON=$(curl -fsSL "$RELEASES_URL") || fail "Could not fetch latest release. Have you published a GitHub release yet?"
elif command -v wget >/dev/null 2>&1; then
  RELEASE_JSON=$(wget -qO- "$RELEASES_URL") || fail "Could not fetch latest release. Have you published a GitHub release yet?"
else
  fail "Neither curl nor wget found. Install one and try again."
fi

TAG=$(printf '%s' "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
[ -z "$TAG" ] && fail "Could not determine latest release tag"

VERSION=$(echo "$TAG" | sed 's/^fragletc-//')
info "Latest version: ${BOLD}${VERSION}${RESET}"

# --- Build asset URL ---
ASSET_NAME="${BINARY}-${PLATFORM}"
[ "$OS" = "windows" ] && ASSET_NAME="${ASSET_NAME}.exe"

DOWNLOAD_BASE="https://github.com/${REPO}/releases/download/${TAG}"
ASSET_URL="${DOWNLOAD_BASE}/${ASSET_NAME}"
CHECKSUMS_URL="${DOWNLOAD_BASE}/checksums.txt"

# --- Download binary ---
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "Downloading ${BOLD}${ASSET_NAME}${RESET}..."

if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${TMPDIR}/${ASSET_NAME}" "$ASSET_URL" || fail "Download failed: ${ASSET_URL}"
  curl -fsSL -o "${TMPDIR}/checksums.txt" "$CHECKSUMS_URL" || warn "Could not download checksums — skipping verification"
else
  wget -qO "${TMPDIR}/${ASSET_NAME}" "$ASSET_URL" || fail "Download failed: ${ASSET_URL}"
  wget -qO "${TMPDIR}/checksums.txt" "$CHECKSUMS_URL" || warn "Could not download checksums — skipping verification"
fi

# --- Verify checksum ---
if [ -f "${TMPDIR}/checksums.txt" ]; then
  EXPECTED=$(grep "${ASSET_NAME}" "${TMPDIR}/checksums.txt" | awk '{print $1}')
  if [ -n "$EXPECTED" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      ACTUAL=$(sha256sum "${TMPDIR}/${ASSET_NAME}" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
      ACTUAL=$(shasum -a 256 "${TMPDIR}/${ASSET_NAME}" | awk '{print $1}')
    else
      warn "No sha256sum or shasum found — skipping checksum verification"
      ACTUAL="$EXPECTED"
    fi
    if [ "$EXPECTED" != "$ACTUAL" ]; then
      fail "Checksum mismatch!\n  Expected: ${EXPECTED}\n  Actual:   ${ACTUAL}"
    fi
    ok "Checksum verified"
  fi
fi

# --- Install ---
mkdir -p "$INSTALL_DIR"
mv "${TMPDIR}/${ASSET_NAME}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"
ok "Installed ${BOLD}${BINARY}${RESET} to ${INSTALL_DIR}/${BINARY}"

# --- PATH check ---
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    warn "${INSTALL_DIR} is not in your PATH"
    printf "\n"
    printf "  Add it to your shell profile:\n"
    printf "\n"
    printf "    ${BOLD}export PATH=\"${INSTALL_DIR}:\$PATH\"${RESET}\n"
    printf "\n"
    ;;
esac

# --- Docker check ---
printf "\n"
if command -v docker >/dev/null 2>&1; then
  ok "Docker found: $(docker --version 2>/dev/null | head -1)"
else
  warn "Docker not found"
  printf "\n"
  printf "  fragletc requires Docker to run code in containers.\n"
  printf "  Install Docker: ${BOLD}https://docs.docker.com/get-docker/${RESET}\n"
  printf "\n"
fi

# --- Success ---
printf "\n"
printf "${GREEN}${BOLD}  Ready to go!${RESET}\n"
printf "\n"
printf "  ${DIM}Run code:${RESET}\n"
printf "    ${BOLD}fragletc --vein=python -c 'print(\"hello\")'${RESET}\n"
printf "\n"
printf "  ${DIM}Start MCP server:${RESET}\n"
printf "    ${BOLD}fragletc mcp${RESET}\n"
printf "\n"
printf "  ${DIM}Full setup guide:${RESET}\n"
printf "    https://github.com/${REPO}/blob/main/INSTALL.md\n"
printf "\n"
