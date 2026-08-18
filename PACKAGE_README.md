# SafeNote Package Building

The `scripts/` directory contains scripts to package SafeNote into `.deb` and `.rpm` formats.

## Prerequisites

### For All Scripts
- Built binary at `build/bin/safenote`
- Icon at `build/linux/safenote.png`
- Desktop file at `build/linux/safenote.desktop`

### Option 1: Using `scripts/package.sh` (Recommended - Uses FPM)
The unified script uses [fpm](https://github.com/jordansissel/fpm) which can build both formats.

**Install fpm:**
```bash
sudo apt-get update
sudo apt-get install -y ruby ruby-dev rubygems build-essential
sudo gem install fpm
```

### Option 2: Using Native Tools
- `scripts/package-deb.sh` - Uses `dpkg-deb` (comes with Debian/Ubuntu)
- `scripts/package-rpm.sh` - Uses `rpmbuild` (install with `sudo dnf install rpm-build`)

## Usage

### Unified Script (scripts/package.sh)

Build both packages:
```bash
./scripts/package.sh
# or from any directory:
/path/to/SafeNote/scripts/package.sh
```

Build only .deb:
```bash
./scripts/package.sh deb
```

Build only .rpm:
```bash
./scripts/package.sh rpm
```

### Individual Scripts

**Build .deb only:**
```bash
./scripts/package-deb.sh
```

**Build .rpm only:**
```bash
./scripts/package-rpm.sh
```

## Output

Packages will be created in the `dist/` directory:
- `.deb`: `safenote_<version>_amd64.deb`
- `.rpm`: `safenote-<version>-1.x86_64.rpm`

## Package Details

### Dependencies
- **.deb**: `libwebkit2gtk-4.1-0`
- **.rpm**: `webkit2gtk-4.1`

### Installation Paths
- Binary: `/usr/bin/safenote`
- Icon: `/usr/share/icons/hicolor/256x256/apps/safenote.png`
- Desktop: `/usr/share/applications/safenote.desktop`

## Installing the Packages

### Debian/Ubuntu (.deb)
```bash
sudo dpkg -i dist/safenote_<version>_amd64.deb
sudo apt-get install -f  # Install dependencies if needed
```

### Fedora/RHEL/CentOS (.rpm)
```bash
sudo dnf install dist/safenote-<version>-1.x86_64.rpm
# or
sudo yum install dist/safenote-<version>-1.x86_64.rpm
```

## Version

The package version is automatically read from `wails.json` under `Info.productVersion`.
