#!/usr/bin/env bash
# Build SafeNote for the current platform, or all platforms with --all.
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v wails >/dev/null 2>&1; then
  echo "wails CLI not found. Install with:"
  echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"
  exit 1
fi

if [[ "${1:-}" == "--all" ]]; then
  echo "Building for windows/amd64..."
  wails build -platform windows/amd64
  echo "Building for linux/amd64..."
  wails build -platform linux/amd64
  echo "Building for darwin/universal..."
  wails build -platform darwin/universal
else
  wails build
fi

echo "Done. Binaries are in build/bin/"
