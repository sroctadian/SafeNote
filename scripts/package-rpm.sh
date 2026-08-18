#!/bin/bash

# SafeNote .rpm Packaging Script (Native rpmbuild)
# Creates .rpm package using rpmbuild

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
ICON_PATH="$PROJECT_ROOT/build/linux/${APP_NAME}.png"
DESKTOP_PATH="$PROJECT_ROOT/build/linux/${APP_NAME}.desktop"
OUTPUT_DIR="$PROJECT_ROOT/dist"

# Dependencies
REQUIRES="webkit2gtk-4.1"

# Check if required files exist
check_files() {
    local missing=0

    if [ ! -f "$BINARY_PATH" ]; then
        echo -e "${RED}Error: Binary not found at $BINARY_PATH${NC}"
        missing=1
    fi

    if [ ! -f "$ICON_PATH" ]; then
        echo -e "${RED}Error: Icon not found at $ICON_PATH${NC}"
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

# Check for rpmbuild
if ! command -v rpmbuild &> /dev/null; then
    echo -e "${YELLOW}Warning: rpmbuild not found. Installing rpm-build...${NC}"
    sudo dnf install -y rpm-build || sudo yum install -y rpm-build
fi

echo -e "${GREEN}Building .rpm package (version $VERSION)...${NC}"

# Setup rpmbuild directory structure
RPM_BUILD_DIR="$HOME/rpmbuild"
mkdir -p "$RPM_BUILD_DIR"/{SOURCES,SPECS,RPMS,SRPMS,BUILD,BUILDROOT}

# Copy source files
SOURCE_TAR="/tmp/${APP_NAME}-${VERSION}.tar.gz"
SOURCE_DIR="/tmp/${APP_NAME}-${VERSION}"

rm -rf "$SOURCE_DIR"
mkdir -p "$SOURCE_DIR"

cp "$BINARY_PATH" "$SOURCE_DIR/$APP_NAME"
cp "$ICON_PATH" "$SOURCE_DIR/$APP_NAME.png"
cp "$DESKTOP_PATH" "$SOURCE_DIR/$APP_NAME.desktop"

# Create source tar (save and restore current directory)
CURRENT_DIR=$(pwd)
cd /tmp
tar -czf "$SOURCE_TAR" "${APP_NAME}-${VERSION}"
cd "$CURRENT_DIR"

# Move source to rpmbuild
mv "$SOURCE_TAR" "$RPM_BUILD_DIR/SOURCES/"

# Create spec file
SPEC_FILE="$RPM_BUILD_DIR/SPECS/${APP_NAME}.spec"
cat > "$SPEC_FILE" << EOF
Name:           ${APP_NAME}
Version:        ${VERSION}
Release:        1%{?dist}
Summary:        Offline, encrypted note-taking desktop app

License:        MIT
URL:            https://github.com/sroctadian/safenote
Source0:        %{name}-%{version}.tar.gz

Requires:       ${REQUIRES}

%description
SafeNote is a secure, offline note-taking application that keeps
your data encrypted and private.

%prep
%setup -q

%build
# No build step, binary is already compiled

%install
rm -rf %{buildroot}

install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_datadir}/icons/hicolor/256x256/apps
install -d %{buildroot}%{_datadir}/applications

install -m 755 %{name} %{buildroot}%{_bindir}/%{name}
install -m 644 %{name}.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/%{name}.png
install -m 644 %{name}.desktop %{buildroot}%{_datadir}/applications/%{name}.desktop

%files
%{_bindir}/%{name}
%{_datadir}/icons/hicolor/256x256/apps/%{name}.png
%{_datadir}/applications/%{name}.desktop

%changelog
* $(date +'%a %b %d %Y') sroctadian <sroctadian> - ${VERSION}-1
- Initial package
EOF

# Build the RPM
rpmbuild -ba "$SPEC_FILE"

# Copy the built RPM to output directory
find "$RPM_BUILD_DIR/RPMS" -name "${APP_NAME}*.rpm" -exec cp {} "$OUTPUT_DIR/" \;

# Cleanup
rm -rf "$SOURCE_DIR"

# Find and display the created package
RPM_FILE=$(find "$OUTPUT_DIR" -name "${APP_NAME}-${VERSION}-1*.rpm" | head -1)
if [ -n "$RPM_FILE" ]; then
    echo -e "${GREEN}✓ .rpm package created: $RPM_FILE${NC}"
else
    echo -e "${RED}Error: Could not find built RPM file${NC}"
    exit 1
fi
