#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MANAGER_HOME="$HOME/.openclaw-manager"
BIN_PATH="$MANAGER_HOME/managerd"
CONFIG_PATH="$MANAGER_HOME/config.toml"
PLIST_PATH="$HOME/Library/LaunchAgents/openclaw-manager.plist"
LABEL="openclaw-manager"

if [[ ! -x "$BIN_PATH" ]]; then
  echo "[ERROR] managerd not found: $BIN_PATH" >&2
  exit 1
fi
if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "[ERROR] config.toml not found: $CONFIG_PATH" >&2
  exit 1
fi

mkdir -p "$HOME/Library/LaunchAgents"
cat > "$PLIST_PATH" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>$LABEL</string>
  <key>ProgramArguments</key>
  <array>
    <string>$BIN_PATH</string>
    <string>--config</string>
    <string>$CONFIG_PATH</string>
    <string>--static-dir</string>
    <string>$ROOT_DIR/src/frontend/dist</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>$MANAGER_HOME/manager.out.log</string>
  <key>StandardErrorPath</key><string>$MANAGER_HOME/manager.err.log</string>
</dict>
</plist>
EOF

launchctl bootout "gui/$(id -u)/$LABEL" >/dev/null 2>&1 || true
launchctl bootstrap "gui/$(id -u)" "$PLIST_PATH"
launchctl kickstart -k "gui/$(id -u)/$LABEL"

launchctl print "gui/$(id -u)/$LABEL" | head -n 40

echo "launchd service installed: $LABEL"
