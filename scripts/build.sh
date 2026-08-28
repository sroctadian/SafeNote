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
  wails build -tags webkit2_41 -platform linux/amd64
  echo "Building for darwin/universal..."
  wails build -platform darwin/universal
else
  wails build -tags webkit2_41
fi

echo "Done. Binaries are in build/bin/"

# Deploy icon and desktop file
mkdir -p ~/.local/bin; cp build/bin/safenote ~/.local/bin/

# Install icons into the hicolor fallback theme. A bare PNG in
# ~/.local/share/icons is NOT a valid icon-theme location and is ignored by
# desktop shells. Regenerate the set after artwork changes:
#   python3 scripts/make-icons.py
rm -f ~/.local/share/icons/safenote.png
if [ -d build/linux/icons/hicolor ]; then
  for icon in build/linux/icons/hicolor/*/apps/safenote.png; do
    size_dir="$(basename "$(dirname "$(dirname "$icon")")")"
    mkdir -p ~/.local/share/icons/hicolor/"$size_dir"/apps
    cp "$icon" ~/.local/share/icons/hicolor/"$size_dir"/apps/safenote.png
  done
  gtk-update-icon-cache -f -t ~/.local/share/icons/hicolor 2>/dev/null || true
else
  echo "warning: build/linux/icons/hicolor missing - run scripts/make-icons.py" >&2
fi
mkdir -p ~/.local/share/applications; cp build/linux/safenote.desktop ~/.local/share/applications/
