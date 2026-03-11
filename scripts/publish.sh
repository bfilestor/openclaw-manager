#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PUBLIC_DIR="$ROOT_DIR/public"
PACKAGE_NAME="openclaw-manager"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
BUILD_DIR="$(mktemp -d)"
RELEASE_DIR="$BUILD_DIR/${PACKAGE_NAME}"
ARCHIVE_NAME="${PACKAGE_NAME}-${TIMESTAMP}.tar.gz"
ARCHIVE_PATH="$PUBLIC_DIR/$ARCHIVE_NAME"

MANAGERD_SRC="$ROOT_DIR/src/bin/managerd"
CONFIG_SRC="$ROOT_DIR/src/bin/config.toml"
DIST_SRC="$ROOT_DIR/src/frontend/dist"
INSTALL_SRC="$ROOT_DIR/scripts/install.sh"
SERVICE_SRC="$ROOT_DIR/openclaw-manager.service"

for path in "$MANAGERD_SRC" "$CONFIG_SRC" "$DIST_SRC" "$INSTALL_SRC" "$SERVICE_SRC"; do
  if [[ ! -e "$path" ]]; then
    echo "[ERROR] Missing required file: $path" >&2
    exit 1
  fi
done

trap 'rm -rf "$BUILD_DIR"' EXIT

mkdir -p "$RELEASE_DIR"/{bin,web,config,scripts,service}
mkdir -p "$PUBLIC_DIR"

install -m 0755 "$MANAGERD_SRC" "$RELEASE_DIR/bin/managerd"
install -m 0644 "$CONFIG_SRC" "$RELEASE_DIR/config/config.toml"
cp -a "$DIST_SRC" "$RELEASE_DIR/web"
install -m 0755 "$INSTALL_SRC" "$RELEASE_DIR/scripts/install.sh"
install -m 0644 "$SERVICE_SRC" "$RELEASE_DIR/service/openclaw-manager.service"

cat > "$RELEASE_DIR/README.txt" <<'EOF'
openclaw-manager release package

Layout:
- bin/managerd
- web/
- config/config.toml
- scripts/install.sh
- service/openclaw-manager.service

deployment guide:
	
cd ~
mkdir ~/.openclaw-manager
tar -xzf openclaw-manager-xxxx.tar.gz -C ~/.openclaw-manager  --strip-components=1
cd ~/.openclaw-manager
chmod +x ./scripts/install.sh
./scripts/install.sh

visit http://127.0.0.1:18799
EOF

(
  cd "$BUILD_DIR"
  tar -czf "$ARCHIVE_PATH" "$PACKAGE_NAME"
)

sha256sum "$ARCHIVE_PATH" > "$ARCHIVE_PATH.sha256"

echo "[OK] Release created: $ARCHIVE_PATH"
echo "[OK] Checksum file: $ARCHIVE_PATH.sha256"
