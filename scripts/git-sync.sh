#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REMOTE="origin"
BRANCH=""
COMMIT_MSG=""

usage() {
  cat <<'EOF'
Usage: git-sync.sh -m "commit message" [-b branch] [-r remote]

Options:
  -m, --message   Commit message (required)
  -b, --branch    Target branch (default: current branch)
  -r, --remote    Remote name (default: origin)
  -h, --help      Show this help message

Behavior:
  1) git add -A
  2) git commit
  3) git push <remote> <branch>
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    -m|--message)
      if [ $# -lt 2 ]; then
        echo "Missing value for $1" >&2
        exit 1
      fi
      COMMIT_MSG="$2"
      shift 2
      ;;
    -b|--branch)
      if [ $# -lt 2 ]; then
        echo "Missing value for $1" >&2
        exit 1
      fi
      BRANCH="$2"
      shift 2
      ;;
    -r|--remote)
      if [ $# -lt 2 ]; then
        echo "Missing value for $1" >&2
        exit 1
      fi
      REMOTE="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [ -z "$COMMIT_MSG" ]; then
  echo "Commit message is required." >&2
  usage >&2
  exit 1
fi

cd "$PROJECT_ROOT"

if [ -z "$BRANCH" ]; then
  BRANCH="$(git rev-parse --abbrev-ref HEAD)"
fi

if [ "$BRANCH" = "HEAD" ]; then
  echo "Detached HEAD detected. Please specify branch with -b." >&2
  exit 1
fi

if [ -z "$(git status --porcelain)" ]; then
  echo "No local changes to commit."
  exit 0
fi

echo "[git] add"
git add -A

echo "[git] commit"
git commit -m "$COMMIT_MSG"

echo "[git] push to $REMOTE/$BRANCH"
git push "$REMOTE" "$BRANCH"

echo "Git sync completed."
