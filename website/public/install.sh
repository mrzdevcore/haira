#!/bin/sh
set -e

REPO="mrzdevcore/haira"
INSTALL_DIR="$HOME/.local/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
echo "Detecting latest Haira release..."
VERSION=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
  echo "Error: could not detect latest version"
  exit 1
fi
echo "Latest version: $VERSION"

# Download
ARCHIVE="haira-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
CHECKSUM_URL="${URL}.sha256"

echo "Downloading $ARCHIVE..."
TMPDIR=$(mktemp -d)
curl -sSL "$URL" -o "$TMPDIR/$ARCHIVE"
curl -sSL "$CHECKSUM_URL" -o "$TMPDIR/$ARCHIVE.sha256"

# Verify checksum
echo "Verifying checksum..."
cd "$TMPDIR"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum -c "$ARCHIVE.sha256"
elif command -v shasum >/dev/null 2>&1; then
  shasum -a 256 -c "$ARCHIVE.sha256"
else
  echo "Warning: no checksum tool found, skipping verification"
fi

# Extract
echo "Installing..."
tar xzf "$ARCHIVE"
EXTRACTED="haira-${VERSION}-${OS}-${ARCH}"

# Install binary (runtime is embedded — no separate runtime directory needed)
mkdir -p "$INSTALL_DIR"
cp "$EXTRACTED/bin/haira" "$INSTALL_DIR/haira"
chmod +x "$INSTALL_DIR/haira"

# Cleanup
rm -rf "$TMPDIR"

echo ""
echo "Haira $VERSION installed successfully!"
echo "  Binary: $INSTALL_DIR/haira"

# Check PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo ""
    echo "Add ~/.local/bin to your PATH:"
    echo "  export PATH=\"\$PATH:\$HOME/.local/bin\""
    echo ""
    echo "Add this line to your ~/.bashrc, ~/.zshrc, or ~/.profile"
    ;;
esac
