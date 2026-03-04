#!/usr/bin/env bash
set -euo pipefail

SERVICE_SRC="$(cd "$(dirname "$0")/.." && pwd)/openclaw-manager.service"
SERVICE_DST="$HOME/.config/systemd/user/openclaw-manager.service"

mkdir -p "$HOME/.config/systemd/user"
cp "$SERVICE_SRC" "$SERVICE_DST"

systemctl --user daemon-reload
systemctl --user enable openclaw-manager.service
systemctl --user restart openclaw-manager.service
systemctl --user status openclaw-manager.service --no-pager

echo "openclaw-manager service installed"
