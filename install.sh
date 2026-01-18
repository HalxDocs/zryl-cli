#!/usr/bin/env sh
set -e

REPO="HalxDocs/zryl-cli"
BIN_NAME="zryl"
INSTALL_DIR="/usr/local/bin"

echo "Installing zryl..."

OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

URL="https://github.com/$REPO/releases/latest/download/zryl-$OS-$ARCH"

TMP="$(mktemp)"
curl -fsSL "$URL" -o "$TMP"

chmod +x "$TMP"

if [ ! -w "$INSTALL_DIR" ]; then
  echo "Requesting sudo permission to install to $INSTALL_DIR"
  sudo mv "$TMP" "$INSTALL_DIR/$BIN_NAME"
else
  mv "$TMP" "$INSTALL_DIR/$BIN_NAME"
fi

echo "zryl installed successfully ✅"
echo "Run: zryl run npm run dev"