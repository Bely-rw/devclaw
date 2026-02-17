#!/usr/bin/env bash
# DevClaw installer — macOS + Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/jholhewres/devclaw/main/scripts/install/install.sh | bash
set -euo pipefail

REPO="jholhewres/devclaw"
BINARY="devclaw"
INSTALL_DIR="/usr/local/bin"
FALLBACK_DIR="$HOME/.local/bin"

info()  { printf '\033[1;34m[info]\033[0m  %s\n' "$1"; }
ok()    { printf '\033[1;32m[ok]\033[0m    %s\n' "$1"; }
err()   { printf '\033[1;31m[error]\033[0m %s\n' "$1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      err "Unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             err "Unsupported architecture: $ARCH" ;;
  esac

  info "Detected: ${OS}/${ARCH}"
}

# Get the latest release tag from GitHub
get_latest_version() {
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')

  if [ -z "$VERSION" ]; then
    err "Could not determine latest version. Check https://github.com/${REPO}/releases"
  fi

  info "Latest version: ${VERSION}"
}

# Download and install
install() {
  ARCHIVE="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
  CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

  TMPDIR=$(mktemp -d)
  trap 'rm -rf "$TMPDIR"' EXIT

  info "Downloading ${URL}..."
  curl -fsSL -o "${TMPDIR}/${ARCHIVE}" "$URL" || err "Download failed"

  # Verify checksum if available
  if curl -fsSL -o "${TMPDIR}/checksums.txt" "$CHECKSUM_URL" 2>/dev/null; then
    EXPECTED=$(grep "$ARCHIVE" "${TMPDIR}/checksums.txt" | awk '{print $1}')
    if [ -n "$EXPECTED" ]; then
      ACTUAL=$(sha256sum "${TMPDIR}/${ARCHIVE}" 2>/dev/null || shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')
      ACTUAL=$(echo "$ACTUAL" | awk '{print $1}')
      if [ "$EXPECTED" != "$ACTUAL" ]; then
        err "Checksum mismatch! Expected: ${EXPECTED}, Got: ${ACTUAL}"
      fi
      ok "Checksum verified"
    fi
  fi

  info "Extracting..."
  tar -xzf "${TMPDIR}/${ARCHIVE}" -C "${TMPDIR}"

  # Install binary
  if [ -w "$INSTALL_DIR" ]; then
    TARGET="$INSTALL_DIR"
  elif command -v sudo &>/dev/null; then
    TARGET="$INSTALL_DIR"
    info "Installing to ${TARGET} (requires sudo)..."
    sudo install -m 755 "${TMPDIR}/${BINARY}" "${TARGET}/${BINARY}"
    ok "Installed to ${TARGET}/${BINARY}"
    return
  else
    TARGET="$FALLBACK_DIR"
    mkdir -p "$TARGET"
    info "No sudo available, installing to ${TARGET}"
  fi

  install -m 755 "${TMPDIR}/${BINARY}" "${TARGET}/${BINARY}"
  ok "Installed to ${TARGET}/${BINARY}"

  # Check if in PATH
  if ! command -v "$BINARY" &>/dev/null; then
    echo ""
    info "Add ${TARGET} to your PATH:"
    echo "  export PATH=\"${TARGET}:\$PATH\""
    echo ""
  fi
}

main() {
  echo ""
  echo "  ╔══════════════════════════════════════╗"
  echo "  ║   DevClaw — AI Agent for Tech Teams  ║"
  echo "  ╚══════════════════════════════════════╝"
  echo ""

  detect_platform
  get_latest_version
  install

  echo ""
  ok "DevClaw installed! Run:"
  echo ""
  echo "  devclaw serve    # start + setup wizard"
  echo "  devclaw --help   # see all commands"
  echo ""
}

main
