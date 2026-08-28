#!/bin/bash

# SafeNote .deb Packaging Script (Native dpkg-deb)
# Creates .deb package using dpkg-deb

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Get version from wails.json
VERSION=$(grep -oP '"productVersion":\s*"\K[^"]+' "$PROJECT_ROOT/wails.json")
if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Could not determine version from wails.json${NC}"
    exit 1
fi

# Configuration (all paths relative to PROJECT_ROOT)
APP_NAME="safenote"
BINARY_PATH="$PROJECT_ROOT/build/bin/${APP_NAME}"
ICON_DIR="$PROJECT_ROOT/build/linux/icons/hicolor"
DESKTOP_PATH="$PROJECT_ROOT/build/linux/${APP_NAME}.desktop"
OUTPUT_DIR="$PROJECT_ROOT/dist"
BUILD_DIR="$PROJECT_ROOT/tmp/deb-build"

# Dependencies
DEPENDS="libgtk-3-0, libwebkit2gtk-4.1-0"

# Check if required files exist
check_files() {
    local missing=0

    if [ ! -f "$BINARY_PATH" ]; then
        echo -e "${RED}Error: Binary not found at $BINARY_PATH${NC}"
        missing=1
    fi

    if [ ! -d "$ICON_DIR" ]; then
        echo -e "${RED}Error: Icon set not found at $ICON_DIR (run scripts/make-icons.py)${NC}"
        missing=1
    fi

    if [ ! -f "$DESKTOP_PATH" ]; then
        echo -e "${RED}Error: Desktop file not found at $DESKTOP_PATH${NC}"
        missing=1
    fi

    if [ $missing -eq 1 ]; then
        exit 1
    fi
}

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

echo -e "${GREEN}Building .deb package (version $VERSION)...${NC}"

# Create directory structure
mkdir -p "$BUILD_DIR/DEBIAN"
mkdir -p "$BUILD_DIR/usr/bin"
mkdir -p "$BUILD_DIR/usr/share/icons/hicolor"
mkdir -p "$BUILD_DIR/usr/share/applications"

# Copy files
cp "$BINARY_PATH" "$BUILD_DIR/usr/bin/$APP_NAME"
chmod +x "$BUILD_DIR/usr/bin/$APP_NAME"

cp -r "$ICON_DIR/." "$BUILD_DIR/usr/share/icons/hicolor/"

cp "$DESKTOP_PATH" "$BUILD_DIR/usr/share/applications/$APP_NAME.desktop"

# Create control file
cat > "$BUILD_DIR/DEBIAN/control" << EOF
Package: $APP_NAME
Version: $VERSION
Architecture: amd64
Maintainer: sroctadian
Depends: $DEPENDS
Section: utils
Priority: optional
Homepage: https://github.com/sroctadian/safenote
Description: Offline, encrypted note-taking desktop app.
 SafeNote is a secure, offline note-taking application that keeps
 your data encrypted and private.
EOF

# Calculate installed size
INSTALLED_SIZE=$(du -sk "$BUILD_DIR" | cut -f1)
echo "Installed-Size: $INSTALLED_SIZE" >> "$BUILD_DIR/DEBIAN/control"

# Create md5sums
(cd "$BUILD_DIR" && find usr -type f -exec md5sum {} \;) > "$BUILD_DIR/DEBIAN/md5sums"

# Build the package
dpkg-deb --build "$BUILD_DIR" "$OUTPUT_DIR/${APP_NAME}_${VERSION}_amd64.deb"

# Cleanup
rm -rf "$BUILD_DIR"

echo -e "${GREEN}✓ .deb package created: $OUTPUT_DIR/${APP_NAME}_${VERSION}_amd64.deb${NC}"
