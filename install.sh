#!/usr/bin/env bash
set -euo pipefail

REPO="edd-framework/edd-core"
VERSION="${EDD_VERSION:-latest}"
BIN_DIR="${EDD_INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "edd: unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux)   EXT="" ;;
    darwin)  EXT="" ;;
    mingw*|msys*|cygwin*) OS="windows"; EXT=".exe" ;;
    *) echo "edd: unsupported OS: $OS"; exit 1 ;;
esac

BINARY="edd-${OS}-${ARCH}${EXT}"

if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
else
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"
fi

echo "edd: installing ${VERSION} for ${OS}/${ARCH}..."

if command -v curl &> /dev/null; then
    curl -sSL "$URL" -o "/tmp/${BINARY}"
elif command -v wget &> /dev/null; then
    wget -q "$URL" -O "/tmp/${BINARY}"
else
    echo "edd: need curl or wget to download"; exit 1
fi

chmod +x "/tmp/${BINARY}"

if [ -w "$BIN_DIR" ]; then
    mv "/tmp/${BINARY}" "${BIN_DIR}/edd${EXT}"
else
    echo "edd: need sudo to install to ${BIN_DIR}"
    sudo mv "/tmp/${BINARY}" "${BIN_DIR}/edd${EXT}"
fi

echo "edd: installed to ${BIN_DIR}/edd${EXT}"
echo "edd: run 'edd help' to get started"
