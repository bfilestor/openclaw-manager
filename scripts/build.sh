#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
FRONTEND_DIR="$SRC_DIR/frontend"
TARGET_BIN="$HOME/.openclaw-manager/managerd"
SERVICE_NAME="openclaw-manager.service"

BUILD_FRONTEND=false
BUILD_BACKEND=false

usage() {
  cat <<'EOF'
Usage: build.sh [--front | --backend | --all]

Options:
  --front     Build frontend only
  --backend   Build backend only
  --all       Build frontend and backend (default)
  -h, --help  Show this help message

Notes:
  - If no option is provided, --all is used.
  - After build, script copies managerd to ~/.openclaw-manager/managerd and restarts service.
EOF
}

for arg in "$@"; do
  case "$arg" in
    --front)
      BUILD_FRONTEND=true
      ;;
    --backend)
      BUILD_BACKEND=true
      ;;
    --all)
      BUILD_FRONTEND=true
      BUILD_BACKEND=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      usage >&2
      exit 1
      ;;
  esac
done

# Default behavior: build everything.
if [ "$BUILD_FRONTEND" = false ] && [ "$BUILD_BACKEND" = false ]; then
  BUILD_FRONTEND=true
  BUILD_BACKEND=true
fi

cd "$PROJECT_ROOT"
git pull

if [ "$BUILD_BACKEND" = true ]; then
  echo "[build] backend"
  cd "$SRC_DIR"
  go clean -cache
  make build
fi

if [ "$BUILD_FRONTEND" = true ]; then
  echo "[build] frontend"
  cd "$FRONTEND_DIR"
  pnpm run build
fi

if [ ! -f "$SRC_DIR/bin/managerd" ]; then
  echo "Backend binary not found: $SRC_DIR/bin/managerd" >&2
  echo "Please run with --backend or --all at least once." >&2
  exit 1
fi

cp "$SRC_DIR/bin/managerd" "$TARGET_BIN"
systemctl --user restart "$SERVICE_NAME"

echo "Build and deploy completed (service restarted)."
