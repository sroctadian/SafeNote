#!/bin/bash

# SafeNote Packaging Script
# Creates .deb and .rpm packages from build artifacts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# Dependencies
DEB_DEPENDS="libgtk-3-0, libwebkit2gtk-4.1-0"
RPM_DEPENDS="gtk3, webkit2gtk-4.1"

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

# Change to project root for fpm to work correctly
cd "$PROJECT_ROOT"

# Check for packaging tools
check_tools() {
    if ! command -v fpm &> /dev/null; then
        echo -e "${YELLOW}Warning: fpm not found. Installing fpm...${NC}"
        sudo apt-get update && sudo apt-get install -y ruby ruby-dev rubygems build-essential
        sudo gem install fpm
    fi
}

# Build .deb package
build_deb() {
    echo -e "${GREEN}Building .deb package...${NC}"

    fpm -s dir -t deb \
        -n "$APP_NAME" \
        -v "$VERSION" \
        --description "Offline, encrypted note-taking desktop app." \
        --url "https://github.com/sroctadian/safenote" \
        --maintainer "sroctadian" \
        --license "MIT" \
        --category "Utility" \
        -d "$DEB_DEPENDS" \
        -p "$OUTPUT_DIR/${APP_NAME}_${VERSION}_amd64.deb" \
        "$BINARY_PATH=/usr/bin/$APP_NAME" \
        "$ICON_DIR=/usr/share/icons" \
        "$DESKTOP_PATH=/usr/share/applications/$APP_NAME.desktop"

    echo -e "${GREEN}✓ .deb package created: $OUTPUT_DIR/${APP_NAME}_${VERSION}_amd64.deb${NC}"
}

# Build .rpm package
build_rpm() {
    echo -e "${GREEN}Building .rpm package...${NC}"

    fpm -s dir -t rpm \
        -n "$APP_NAME" \
        -v "$VERSION" \
        --description "Offline, encrypted note-taking desktop app." \
        --url "https://github.com/sroctadian/safenote" \
        --maintainer "sroctadian" \
        --license "MIT" \
        --category "Utility" \
        -d "$RPM_DEPENDS" \
        -p "$OUTPUT_DIR/${APP_NAME}-${VERSION}-1.x86_64.rpm" \
        "$BINARY_PATH=/usr/bin/$APP_NAME" \
        "$ICON_DIR=/usr/share/icons" \
        "$DESKTOP_PATH=/usr/share/applications/$APP_NAME.desktop"

    echo -e "${GREEN}✓ .rpm package created: $OUTPUT_DIR/${APP_NAME}-${VERSION}-1.x86_64.rpm${NC}"
}

# Main execution
main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  SafeNote Packaging Script${NC}"
    echo -e "${GREEN}  Version: $VERSION${NC}"
    echo -e "${GREEN}========================================${NC}"

    check_files
    check_tools

    # Parse arguments
    BUILD_DEB=false
    BUILD_RPM=false

    if [ $# -eq 0 ]; then
        BUILD_DEB=true
        BUILD_RPM=true
    else
        for arg in "$@"; do
            case $arg in
                deb)
                    BUILD_DEB=true
                    ;;
                rpm)
                    BUILD_RPM=true
                    ;;
                all)
                    BUILD_DEB=true
                    BUILD_RPM=true
                    ;;
                *)
                    echo -e "${RED}Unknown argument: $arg${NC}"
                    echo "Usage: $0 [deb|rpm|all]"
                    exit 1
                    ;;
            esac
        done
    fi

    if [ "$BUILD_DEB" = true ]; then
        build_deb
    fi

    if [ "$BUILD_RPM" = true ]; then
        build_rpm
    fi

    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Packaging complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
}

main "$@"
