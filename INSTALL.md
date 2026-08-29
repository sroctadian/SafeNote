# INSTALL

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | 1.24+ | https://go.dev/dl/ |
| Node.js | 18+ | for the frontend build (Vite + Tailwind + DaisyUI) |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Platform build deps | — | see below |

**Linux**: `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` (or `4.1` on newer distros)
**Windows**: WebView2 runtime (bundled with Windows 11; installable on Windows 10)
**macOS**: Xcode Command Line Tools (`xcode-select --install`)

Verify your environment is ready:

```bash
wails doctor
```

## First-time setup

```bash
git clone <this-repo> SafeNote && cd SafeNote
go mod tidy                 # resolves and downloads Go dependencies
cd frontend && npm install && cd ..
```

`go mod tidy` requires network access to the Go module proxy
(`proxy.golang.org` by default, or your configured `GOPROXY`). This step
could not be run in the generation sandbox — run it in your own
environment before building.

## Development (hot reload)

```bash
wails dev
```

This starts the Go backend and a Vite dev server with hot reload for the
frontend, opening a native window pointed at it.

## Running tests

```bash
go test ./... -v -cover
```

Target: ≥80% coverage on `internal/crypto` and `internal/service`
(the security-critical packages). Current test files cover:

- `internal/crypto/crypto_test.go` — KDF determinism, wrong-PIN/tamper
  rejection, wipe behavior, salt/nonce uniqueness
- `internal/service/note_service_test.go` — create/open round trip,
  generic wrong-PIN error, cooldown lockout, edit re-encryption,
  delete/list/search
- `internal/service/backup_service_test.go` — export never leaks
  plaintext, restore duplicate detection, no-overwrite-by-default,
  checksum tamper detection

## Building a release binary

```bash
./scripts/build.sh          # current platform
./scripts/build.sh --all    # windows/amd64, linux/amd64, darwin/universal
```

On Windows: `.\scripts\build.ps1` or `.\scripts\build.ps1 -All`

Output binaries land in `build/bin/`. Add an app icon at
`build/windows/icon.ico`, `build/darwin/icon.icns` before producing a
distributable release — see `build/README.md`.

## Data locations

SafeNote stores its SQLite database and vault key under the OS user
config directory (`os.UserConfigDir()`):

- Windows: `%AppData%\SafeNote\`
- Linux: `~/.config/SafeNote/`
- macOS: `~/Library/Application Support/SafeNote/`

## Troubleshooting

- **`wails: command not found`**: ensure `$(go env GOPATH)/bin` is on
  your `PATH`.
- **CGO errors building `modernc.org/sqlite`**: this driver is pure Go
  (no CGO required); if you see CGO errors, check you haven't
  accidentally switched to `mattn/go-sqlite3` without a C toolchain.
- **Blank window / assets not found**: run `cd frontend && npm run build`
  once so `frontend/dist` exists before `wails build` (it embeds that
  directory via `//go:embed all:frontend/dist` in `main.go`).
