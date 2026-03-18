#!/usr/bin/env bash
set -euo pipefail

LABEL="openclaw-manager"
PLIST_PATH="$HOME/Library/LaunchAgents/openclaw-manager.plist"

launchctl bootout "gui/$(id -u)/$LABEL" >/dev/null 2>&1 || true
rm -f "$PLIST_PATH"

echo "launchd service removed: $LABEL"
