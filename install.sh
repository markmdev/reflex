#!/bin/sh
set -e

REPO="markmdev/reflex"
INSTALL_DIR="$HOME/.local/bin"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

BINARY="reflex-${OS}-${ARCH}"

# Get latest release tag
TAG=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [ -z "$TAG" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${BINARY}"

echo "Downloading reflex ${TAG} (${OS}/${ARCH})..."
mkdir -p "$INSTALL_DIR"
curl -sL "$URL" -o "${INSTALL_DIR}/reflex"
chmod +x "${INSTALL_DIR}/reflex"

echo "Installed reflex ${TAG} to ${INSTALL_DIR}/reflex"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo ""
     echo "Add ${INSTALL_DIR} to your PATH:"
     echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac
