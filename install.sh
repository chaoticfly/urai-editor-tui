#!/usr/bin/env bash
# install.sh — install urai from GitHub Releases
#
#   curl -fsSL https://raw.githubusercontent.com/chaoticfly/urai-editor-tui/master/install.sh | bash
#   wget -qO- https://raw.githubusercontent.com/chaoticfly/urai-editor-tui/master/install.sh | bash
#
# Optional: pin a version
#   curl -fsSL .../install.sh | VERSION=v1.0.2 bash

set -euo pipefail

REPO="chaoticfly/urai-editor-tui"
BIN="urai"
API_URL="https://api.github.com/repos/${REPO}/releases"

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

# ── Resolve tag ───────────────────────────────────────────────────────────────
# Use the GitHub API to get the exact tag; avoids redirect issues with
# /releases/latest/download/ when piping through curl | bash.

_fetch() {
  if command -v curl &>/dev/null; then
    curl -fsSL "$1"
  elif command -v wget &>/dev/null; then
    wget -qO- "$1"
  else
    echo "error: curl or wget is required" >&2
    exit 1
  fi
}

if [[ "${VERSION:-latest}" == "latest" ]]; then
  TAG=$(_fetch "${API_URL}/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
else
  TAG="${VERSION}"
fi

if [[ -z "${TAG}" ]]; then
  echo "error: could not determine latest release tag" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

# ── Pick install directory ────────────────────────────────────────────────────

if [ -w /usr/local/bin ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "${INSTALL_DIR}"
fi

# ── Download & install ────────────────────────────────────────────────────────

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Installing urai ${TAG} ..."
_fetch "${URL}" > "${TMP}/${ASSET}"

tar -xzf "${TMP}/${ASSET}" -C "${TMP}" "${BIN}"
install -m 755 "${TMP}/${BIN}" "${INSTALL_DIR}/${BIN}"

echo "Installed ${INSTALL_DIR}/${BIN}"

# ── PATH hint ─────────────────────────────────────────────────────────────────

if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
  case "${SHELL:-}" in
    */zsh)  SHELL_RC="~/.zshrc" ;;
    */fish) SHELL_RC="~/.config/fish/config.fish" ;;
    *)      SHELL_RC="~/.bashrc" ;;
  esac
  echo ""
  echo "  Add ${INSTALL_DIR} to your PATH — append to ${SHELL_RC}:"
  echo ""
  echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
  echo ""
  echo "  Then: source ${SHELL_RC}"
  echo ""
fi

echo "Done. Run: ${BIN}"
