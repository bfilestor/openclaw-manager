#!/usr/bin/env bash
set -euo pipefail

myhome="${HOME:-}"
if [[ -z "$myhome" ]]; then
  myhome="$(getent passwd "$(id -un)" | cut -d: -f6)"
fi
if [[ -z "$myhome" ]]; then
  echo "[ERROR] cannot resolve current user home directory" >&2
  exit 1
fi

openclaw_path="$(command -v openclaw || true)"
node_path="$(command -v node || true)"

if [[ -z "$openclaw_path" ]]; then
  echo "[ERROR] openclaw not found in PATH" >&2
  exit 1
fi
if [[ -z "$node_path" ]]; then
  echo "[ERROR] node not found in PATH" >&2
  exit 1
fi

openclaw_bin_dir="$(dirname "$openclaw_path")"
node_bin_dir="$(dirname "$node_path")"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_TEMPLATE="$ROOT_DIR/service/openclaw-manager.service"

if [[ ! -f "$SERVICE_TEMPLATE" ]]; then
  echo "[ERROR] service template not found: $SERVICE_TEMPLATE" >&2
  exit 1
fi

sed -i \
  -e "s|\${node_bin_dir}|$node_bin_dir|g" \
  -e "s|\${openclaw_bin_dir}|$openclaw_bin_dir|g" \
  "$SERVICE_TEMPLATE"


SERVICE_DST="$HOME/.config/systemd/user/openclaw-manager.service"
CONFIG_FILE="$ROOT_DIR/config/config.toml"

rand32() {
  LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32
}

mkdir -p "$HOME/.config/systemd/user"
cp "$SERVICE_TEMPLATE" "$SERVICE_DST"

if [[ -f "$CONFIG_FILE" ]]; then
  jwt_secret="$(rand32)"
  reset_super_token="$(rand32)"
  while [[ "$reset_super_token" == "$jwt_secret" ]]; do
    reset_super_token="$(rand32)"
  done

  sed -i \
    -e "s|replace-with-strong-secret-32bytes-min|$jwt_secret|g" \
    -e "s|replace-with-another-strong-secret-32bytes-min|$reset_super_token|g" \
    "$CONFIG_FILE"

  echo "[OK] config secrets initialized: $CONFIG_FILE"
else
  echo "[WARN] config file not found, skip secret initialization: $CONFIG_FILE"
fi

systemctl --user daemon-reload
systemctl --user enable openclaw-manager.service
systemctl --user restart openclaw-manager.service
systemctl --user status openclaw-manager.service --no-pager

echo "openclaw-manager service installed"
