# Build SafeNote on Windows.
# Usage: .\scripts\build.ps1 [-All]
param(
  [switch]$All
)

Set-Location (Join-Path $PSScriptRoot "..")

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
  Write-Error "wails CLI not found. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
  exit 1
}

if ($All) {
  wails build -platform windows/amd64
  wails build -platform linux/amd64
  wails build -platform darwin/universal
} else {
  wails build
}

Write-Host "Done. Binaries are in build/bin/"
