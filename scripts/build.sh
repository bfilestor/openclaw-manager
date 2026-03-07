#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
FRONTEND_DIR="$SRC_DIR/frontend"
TARGET_BIN="$HOME/.openclaw-manager/managerd"
RESTART_SERVICE=true

for arg in "$@"; do
  case "$arg" in
    --no-restart)
      RESTART_SERVICE=false
      ;;
    -h|--help)
      cat <<'EOF'
Usage: build.sh [--no-restart]

Options:
  --no-restart   Build and deploy binary without restarting service
  -h, --help     Show this help message
EOF
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      echo "Use --help to see available options." >&2
      exit 1
      ;;
  esac
done

cd "$PROJECT_ROOT"
git pull

cd "$SRC_DIR"
go clean -cache
make build

cd "$FRONTEND_DIR"
pnpm run build

cp "$SRC_DIR/bin/managerd" "$TARGET_BIN"

if [ "$RESTART_SERVICE" = true ]; then
  systemctl --user restart openclaw-manager.service
  echo "Build and deploy completed (service restarted)."
else
  echo "Build and deploy completed (service not restarted)."
fi
