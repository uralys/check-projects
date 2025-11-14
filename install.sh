#!/bin/sh
set -e

REPO="uralys/check-projects"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="check-projects"

print_banner() {
  echo ""
  echo "╔══════════════════════════════════════════╗"
  echo "║         check-projects installer         ║"
  echo "╚══════════════════════════════════════════╝"
  echo ""
}

print_success() {
  echo "✓ $1"
}

print_error() {
  echo "✗ Error: $1" >&2
  exit 1
}

detect_os() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$OS" in
    linux*)
      OS="Linux"
      ;;
    darwin*)
      OS="Darwin"
      ;;
    *)
      print_error "Unsupported OS: $OS"
      ;;
  esac
}

detect_arch() {
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64)
      ARCH="x86_64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      print_error "Unsupported architecture: $ARCH"
      ;;
  esac
}

get_latest_release() {
  LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

  if [ -z "$LATEST_RELEASE" ]; then
    print_error "Failed to get latest release"
  fi

  print_success "Latest version: $LATEST_RELEASE"
}

download_binary() {
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}_${LATEST_RELEASE}_${OS}_${ARCH}.tar.gz"
  TMP_DIR=$(mktemp -d)
  TMP_FILE="${TMP_DIR}/${BINARY_NAME}.tar.gz"

  echo "Downloading from: $DOWNLOAD_URL"

  if ! curl -fsSL -o "$TMP_FILE" "$DOWNLOAD_URL"; then
    print_error "Download failed"
  fi

  print_success "Downloaded successfully"

  cd "$TMP_DIR"
  tar -xzf "${BINARY_NAME}.tar.gz" || print_error "Failed to extract archive"

  if [ ! -f "$BINARY_NAME" ]; then
    print_error "Binary not found in archive"
  fi

  print_success "Extracted successfully"
}

install_binary() {
  mkdir -p "$INSTALL_DIR"

  if ! mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"; then
    print_error "Failed to install binary to $INSTALL_DIR"
  fi

  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

  rm -rf "$TMP_DIR"

  print_success "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

check_path() {
  case ":$PATH:" in
    *":${INSTALL_DIR}:"*)
      print_success "${INSTALL_DIR} is in your PATH"
      ;;
    *)
      echo ""
      echo "⚠ Warning: ${INSTALL_DIR} is not in your PATH"
      echo ""
      echo "Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
      echo ""
      echo "  export PATH=\"\${HOME}/.local/bin:\${PATH}\""
      echo ""
      ;;
  esac
}

verify_installation() {
  if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
    if [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
      echo ""
      echo "Installation completed but ${BINARY_NAME} is not in your PATH."
      echo "You can run it with: ${INSTALL_DIR}/${BINARY_NAME}"
      return
    else
      print_error "Installation verification failed"
    fi
  fi

  VERSION=$("$BINARY_NAME" --version 2>&1 || echo "unknown")
  echo ""
  print_success "Installation successful! Version: $VERSION"
  echo ""
  echo "Run 'check-projects --help' to get started"
}

main() {
  print_banner
  detect_os
  detect_arch
  print_success "Detected: $OS $ARCH"
  get_latest_release
  download_binary
  install_binary
  check_path
  verify_installation
}

main
