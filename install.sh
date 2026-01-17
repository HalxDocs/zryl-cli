#!/usr/bin/env sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

BINARY="zryl-${OS}-${ARCH}"
URL="https://github.com/HalxDocs/zryl/releases/latest/download/$BINARY"

INSTALL_DIR="/usr/local/bin"

echo "Installing zryl..."
curl -fsSL "$URL" -o "$INSTALL_DIR/zryl"
chmod +x "$INSTALL_DIR/zryl"

echo "zryl installed successfully 🎉"
 