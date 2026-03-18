#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
TAURI_DIR="$ROOT/desktop/tauri"
RES_DIR="$TAURI_DIR/src-tauri/resources"

mkdir -p "$RES_DIR"

if [[ -f "$ROOT/src/bin/managerd" ]]; then
  cp -f "$ROOT/src/bin/managerd" "$RES_DIR/managerd"
fi
if [[ -f "$ROOT/src/bin/managerd.exe" ]]; then
  cp -f "$ROOT/src/bin/managerd.exe" "$RES_DIR/managerd.exe"
fi

rm -rf "$RES_DIR/frontend-dist"
mkdir -p "$RES_DIR/frontend-dist"
cp -r "$ROOT/src/frontend/dist"/* "$RES_DIR/frontend-dist"/

echo "Assets prepared under $RES_DIR"
