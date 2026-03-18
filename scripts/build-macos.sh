#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
FRONTEND_DIR="$SRC_DIR/frontend"
TARGET_BIN="$HOME/.openclaw-manager/managerd"

BUILD_FRONTEND=false
BUILD_BACKEND=false

if [[ $# -eq 0 ]]; then
  BUILD_FRONTEND=true
  BUILD_BACKEND=true
fi

for arg in "$@"; do
  case "$arg" in
    --front) BUILD_FRONTEND=true ;;
    --backend) BUILD_BACKEND=true ;;
    --all) BUILD_FRONTEND=true; BUILD_BACKEND=true ;;
    *) echo "Unknown option: $arg" >&2; exit 1 ;;
  esac
done

cd "$PROJECT_ROOT"

git pull

if [[ "$BUILD_BACKEND" == true ]]; then
  echo "[build] backend"
  cd "$SRC_DIR"
  CGO_ENABLED=0 go build -o bin/managerd ./cmd/server
fi

if [[ "$BUILD_FRONTEND" == true ]]; then
  echo "[build] frontend"
  cd "$FRONTEND_DIR"
  pnpm run build
fi

mkdir -p "$(dirname "$TARGET_BIN")"
cp "$SRC_DIR/bin/managerd" "$TARGET_BIN"

echo "Build completed. Run scripts/install-launchd.sh to (re)load launchd service."
