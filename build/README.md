# build/

This directory holds platform-specific packaging assets generated and
consumed by the Wails CLI:

- `windows/` — `icon.ico`, `info.json`, NSIS installer config
- `linux/` — `.desktop` file, AppImage/deb metadata
- `darwin/` — `icon.icns`, `Info.plist`, entitlements for notarization

Run `wails build` (or `scripts/build.sh` / `scripts/build.ps1`) once to
have the Wails CLI scaffold platform-specific defaults here if they are
not already present, then customize (app icon, bundle identifier,
signing) as needed for your release process. This scaffold intentionally
ships without a binary icon asset; add your own `.ico` / `.icns` /
`.png` before producing a release build.
