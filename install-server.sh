#!/usr/bin/env bash
# install-server.sh — install urai-ssh on a Raspberry Pi (linux/arm64)
#
#   curl -fsSL https://raw.githubusercontent.com/chaoticfly/urai-editor-tui/master/install-server.sh | bash
#   wget -qO- https://raw.githubusercontent.com/chaoticfly/urai-editor-tui/master/install-server.sh | bash
#
# Optional: pin a version
#   curl -fsSL .../install-server.sh | VERSION=v1.0.5 bash
#
# What this script does:
#   1. Downloads urai-ssh linux/arm64 from GitHub Releases
#   2. Installs it to /usr/local/bin/urai-ssh
#   3. Installs a systemd service (urai-ssh@.service)
#   4. Enables and starts the service as the current user on port 2222

set -euo pipefail

REPO="chaoticfly/urai-editor-tui"
BIN="urai-ssh"
ASSET="urai-ssh-linux-arm64.tar.gz"
API_URL="https://api.github.com/repos/${REPO}/releases"
SERVICE_NAME="urai-ssh"
PORT="${PORT:-2222}"

# ── Sanity checks ─────────────────────────────────────────────────────────────

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "error: this script is for Linux (Raspberry Pi) only" >&2
  exit 1
fi

if [[ "$(uname -m)" != "aarch64" && "$(uname -m)" != "arm64" ]]; then
  echo "error: this script only supports arm64 (aarch64) — got $(uname -m)" >&2
  exit 1
fi

if ! command -v systemctl &>/dev/null; then
  echo "error: systemd is required" >&2
  exit 1
fi

# ── Helpers ───────────────────────────────────────────────────────────────────

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

# ── Resolve release tag ───────────────────────────────────────────────────────

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

# ── Download & install binary ─────────────────────────────────────────────────

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Downloading urai-ssh ${TAG} ..."
_fetch "${URL}" > "${TMP}/${ASSET}"
tar -xzf "${TMP}/${ASSET}" -C "${TMP}" "${BIN}"

echo "Installing /usr/local/bin/${BIN} ..."
sudo install -m 755 "${TMP}/${BIN}" "/usr/local/bin/${BIN}"

# ── Install systemd service ───────────────────────────────────────────────────

SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}@.service"
CURRENT_USER="${SUDO_USER:-${USER}}"

echo "Installing systemd service to ${SERVICE_FILE} ..."
sudo tee "${SERVICE_FILE}" > /dev/null <<EOF
[Unit]
Description=urai SSH editor server
After=network.target

[Service]
Type=simple
User=%i
ExecStart=/usr/local/bin/urai-ssh --addr 0.0.0.0:${PORT}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now "${SERVICE_NAME}@${CURRENT_USER}"

# ── Done ──────────────────────────────────────────────────────────────────────

echo ""
echo "urai-ssh ${TAG} is running on port ${PORT}."
echo ""
echo "  Connect from your Tailscale network:"
echo "    ssh -t ${CURRENT_USER}@<pi-hostname> urai-ssh"
echo "    ssh -t ${CURRENT_USER}@<pi-hostname> urai-ssh /path/to/file.txt"
echo ""
echo "  Service management:"
echo "    sudo systemctl status ${SERVICE_NAME}@${CURRENT_USER}"
echo "    sudo journalctl -u ${SERVICE_NAME}@${CURRENT_USER} -f"
echo ""
