#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="openclaw-manager.service"
SERVICE_FILE="$HOME/.config/systemd/user/$SERVICE_NAME"

systemctl --user disable --now "$SERVICE_NAME" 2>/dev/null || true
systemctl --user stop "$SERVICE_NAME" 2>/dev/null || true
systemctl --user reset-failed "$SERVICE_NAME" 2>/dev/null || true

rm -f "$SERVICE_FILE"

systemctl --user daemon-reload

echo "openclaw-manager service uninstalled"
