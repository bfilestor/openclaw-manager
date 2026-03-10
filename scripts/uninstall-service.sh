#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="openclaw-manager.service"
SERVICE_FILE="$HOME/.config/systemd/user/$SERVICE_NAME"

if systemctl --user list-unit-files | grep -q "^${SERVICE_NAME}"; then
  systemctl --user disable --now "$SERVICE_NAME" || true
else
  # Best effort stop in case unit is loaded transiently.
  systemctl --user stop "$SERVICE_NAME" || true
fi

rm -f "$SERVICE_FILE"

systemctl --user daemon-reload
systemctl --user reset-failed "$SERVICE_NAME" || true

echo "openclaw-manager service uninstalled"
