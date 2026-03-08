#!/usr/bin/env bash
# install.sh — install urai from GitHub Releases
#
#   curl -fsSL https://raw.githubusercontent.com/OWNER/urai/master/install.sh | bash
#   wget -qO- https://raw.githubusercontent.com/OWNER/urai/master/install.sh | bash
#
# Optional: pin a version
#   curl -fsSL .../install.sh | VERSION=v0.2.0 bash

set -euo pipefail

REPO="OWNER/urai"   # ← replace OWNER with your GitHub username / org
BIN="urai"
BASE_URL="https://github.com/${REPO}/releases"
VERSION="${VERSION:-latest}"

# ── Detect OS / arch ─────────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"

case "${OS}" in
  Linux)  GOOS="linux"  ;;
  Darwin) GOOS="darwin" ;;
  *)
    echo "error: unsupported OS '${OS}'" >&2
    exit 1
    ;;
esac

case "${ARCH}" in
  x86_64)          GOARCH="amd64" ;;
  aarch64 | arm64) GOARCH="arm64" ;;
  *)
    echo "error: unsupported architecture '${ARCH}'" >&2
    exit 1
    ;;
esac

if [[ "${GOOS}" == "darwin" && "${GOARCH}" == "amd64" ]]; then
  echo "error: no pre-built binary for macOS Intel." >&2
  echo "       Build from source: cd prose && go build ./cmd/urai/" >&2
  exit 1
fi

ASSET="${BIN}-${GOOS}-${GOARCH}.tar.gz"

# ── Pick install directory ────────────────────────────────────────────────────
# Prefer /usr/local/bin (no sudo needed on macOS; writable for root on Linux).
# Fall back to ~/.local/bin for unprivileged Linux users.

if [ -w /usr/local/bin ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "${INSTALL_DIR}"
fi

# ── Download URL ──────────────────────────────────────────────────────────────

if [[ "${VERSION}" == "latest" ]]; then
  URL="${BASE_URL}/latest/download/${ASSET}"
else
  URL="${BASE_URL}/download/${VERSION}/${ASSET}"
fi

# ── Download & install ────────────────────────────────────────────────────────

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Downloading ${ASSET} ..."
if command -v curl &>/dev/null; then
  curl -fsSL "${URL}" -o "${TMP}/${ASSET}"
elif command -v wget &>/dev/null; then
  wget -qO "${TMP}/${ASSET}" "${URL}"
else
  echo "error: curl or wget is required" >&2
  exit 1
fi

tar -xzf "${TMP}/${ASSET}" -C "${TMP}" "${BIN}"
install -m 755 "${TMP}/${BIN}" "${INSTALL_DIR}/${BIN}"

echo "Installed ${INSTALL_DIR}/${BIN}"

# ── PATH hint (only needed for ~/.local/bin fallback) ─────────────────────────

if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
  SHELL_RC=""
  case "${SHELL}" in
    */zsh)  SHELL_RC="~/.zshrc"  ;;
    */fish) SHELL_RC="~/.config/fish/config.fish" ;;
    *)      SHELL_RC="~/.bashrc" ;;
  esac
  echo ""
  echo "  Add ${INSTALL_DIR} to your PATH by adding this to ${SHELL_RC}:"
  echo ""
  echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
  echo ""
  echo "  Then restart your shell or run: source ${SHELL_RC}"
  echo ""
fi

echo "Done. Run: ${BIN}"
