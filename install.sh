#!/bin/sh
set -e

REPO="3000-2/nudge"
INSTALL_DIR="/usr/local/bin"
BINARY="nudge"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { printf "${CYAN}▸${NC} %s\n" "$1"; }
success() { printf "${GREEN}✓${NC} %s\n" "$1"; }
fail() { printf "${RED}✗${NC} %s\n" "$1" >&2; exit 1; }

# macOS only
OS="$(uname -s)"
if [ "$OS" != "Darwin" ]; then
  fail "nudge only supports macOS (got $OS)"
fi

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64)   ARCH="arm64" ;;
  *)       fail "unsupported architecture: $ARCH" ;;
esac

info "Detected: macOS $ARCH"

# Get latest release tag
info "Fetching latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$TAG" ]; then
  fail "could not determine latest release"
fi

info "Latest version: $TAG"

# Download
ARCHIVE="nudge_darwin_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "Downloading ${URL}..."
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}" || fail "download failed"

# Extract
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR" || fail "extract failed"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  info "Requesting sudo to install to ${INSTALL_DIR}..."
  sudo mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

chmod +x "${INSTALL_DIR}/${BINARY}"

success "nudge ${TAG} installed to ${INSTALL_DIR}/${BINARY}"

# Verify
"${INSTALL_DIR}/${BINARY}" version
