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
  -e "s|\${myhome}|$myhome|g" \
  "$SERVICE_TEMPLATE"

SERVICE_SRC="$(cd "$(dirname "$0")/.." && pwd)/service/openclaw-manager.service"
SERVICE_DST="$HOME/.config/systemd/user/openclaw-manager.service"

mkdir -p "$HOME/.config/systemd/user"
cp "$SERVICE_SRC" "$SERVICE_DST"

systemctl --user daemon-reload
systemctl --user enable openclaw-manager.service
systemctl --user restart openclaw-manager.service
systemctl --user status openclaw-manager.service --no-pager

echo "openclaw-manager service installed"
