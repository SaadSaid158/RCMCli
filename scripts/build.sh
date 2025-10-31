#!/bin/bash
# RCMCli Production Build Script
# Builds optimized, stripped binaries for all platforms with CGO support for macOS

set -e

VERSION="1.0.0"
DIST_DIR="dist"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC' 2>/dev/null || date '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags for production
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"
BUILD_FLAGS="-trimpath -ldflags=\"${LDFLAGS}\""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║       RCMCli Production Build Script v${VERSION}           ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "[*] Build Configuration:"
echo "    Version:    ${VERSION}"
echo "    Build Time: ${BUILD_TIME}"
echo "    Git Commit: ${GIT_COMMIT}"
echo "    Flags:      -s -w (stripped & optimized)"
echo "    Trimpath:   Enabled (reproducible builds)"
echo ""

# Initialize dependencies
if [ ! -f "go.sum" ]; then
    echo "[*] Initializing dependencies..."
    go mod download
    go mod tidy
    echo "[+] Dependencies ready"
    echo ""
fi

# Clean previous builds
if [ -d "$DIST_DIR" ]; then
    echo "[*] Cleaning previous builds..."
    rm -rf "$DIST_DIR"
fi

# Create dist directories
mkdir -p "$DIST_DIR/Linux"
mkdir -p "$DIST_DIR/Windows"
mkdir -p "$DIST_DIR/macOS"

echo "[*] Building for all platforms..."
echo ""

# Linux x86_64 (Pure Go)
echo "[*] Building for Linux x86_64 (pure Go)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/Linux/RCMCli_${VERSION}_linux_amd64"
echo "[+] Built: $DIST_DIR/Linux/RCMCli_${VERSION}_linux_amd64"

# Linux ARM64 (Pure Go)
echo "[*] Building for Linux ARM64 (pure Go)..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/Linux/RCMCli_${VERSION}_linux_arm64"
if [ $? -eq 0 ]; then
    echo "[+] Built: $DIST_DIR/Linux/RCMCli_${VERSION}_linux_arm64"
else
    echo "[-] Build failed for Linux ARM64"
fi

# Windows x86_64 (Pure Go)
echo "[*] Building for Windows x86_64 (pure Go)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/Windows/RCMCli_${VERSION}_windows_amd64.exe"
if [ $? -eq 0 ]; then
    echo "[+] Built: $DIST_DIR/Windows/RCMCli_${VERSION}_windows_amd64.exe"
else
    echo "[-] Build failed for Windows x86_64"
fi

# Windows ARM64 (Pure Go)
echo "[*] Building for Windows ARM64 (pure Go)..."
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/Windows/RCMCli_${VERSION}_windows_arm64.exe"
if [ $? -eq 0 ]; then
    echo "[+] Built: $DIST_DIR/Windows/RCMCli_${VERSION}_windows_arm64.exe"
else
    echo "[-] Build failed for Windows ARM64"
fi

# macOS x86_64 (with CGO for IOKit)
echo "[*] Building for macOS x86_64 (with IOKit support)..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # Native macOS build with CGO
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/macOS/RCMCli_${VERSION}_darwin_amd64"
    if [ $? -eq 0 ]; then
        echo "[+] Built: $DIST_DIR/macOS/RCMCli_${VERSION}_darwin_amd64 (native IOKit)"
    else
        echo "[-] Build failed for macOS x86_64"
    fi
else
    # Cross-compile without CGO (fallback mode)
    echo "[!] Not on macOS - building fallback version (requires TegraRcmSmash)"
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/macOS/RCMCli_${VERSION}_darwin_amd64"
    if [ $? -eq 0 ]; then
        echo "[+] Built: $DIST_DIR/macOS/RCMCli_${VERSION}_darwin_amd64 (fallback mode)"
    else
        echo "[-] Build failed for macOS x86_64"
    fi
fi

# macOS ARM64 (Apple Silicon with CGO for IOKit)
echo "[*] Building for macOS ARM64 (with IOKit support)..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # Native macOS build with CGO
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/macOS/RCMCli_${VERSION}_darwin_arm64"
    if [ $? -eq 0 ]; then
        echo "[+] Built: $DIST_DIR/macOS/RCMCli_${VERSION}_darwin_arm64 (native IOKit)"
    else
        echo "[-] Build failed for macOS ARM64"
    fi
else
    # Cross-compile without CGO (fallback mode)
    echo "[!] Not on macOS - building fallback version (requires TegraRcmSmash)"
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o "$DIST_DIR/macOS/RCMCli_${VERSION}_darwin_arm64"
    if [ $? -eq 0 ]; then
        echo "[+] Built: $DIST_DIR/macOS/RCMCli_${VERSION}_darwin_arm64 (fallback mode)"
    else
        echo "[-] Build failed for macOS ARM64"
    fi
fi

echo ""
echo "[*] Generating checksums..."
cd "$DIST_DIR"
find . -type f -name "RCMCli_*" -exec sha256sum {} \; > checksums.txt 2>/dev/null || find . -type f -name "RCMCli_*" -exec shasum -a 256 {} \; > checksums.txt
cd ..

echo "[*] Calculating binary sizes..."
echo ""
printf "%-50s %10s\n" "Binary" "Size"
printf "%-50s %10s\n" "------" "----"
find "$DIST_DIR" -type f -name "RCMCli_*" | while read file; do
    size=$(du -h "$file" | cut -f1)
    basename=$(basename "$file")
    printf "%-50s %10s\n" "$basename" "$size"
done

echo ""
echo "[+] All builds complete!"
echo ""
echo "Build artifacts:"
find "$DIST_DIR/" -type f | sort
echo ""
echo "Checksums:"
cat "$DIST_DIR/checksums.txt"
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                  Build Summary                             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo "  Version:     ${VERSION}"
echo "  Build Time:  ${BUILD_TIME}"
echo "  Git Commit:  ${GIT_COMMIT}"
echo "  Output Dir:  ${DIST_DIR}/"
echo ""
echo "[*] Production builds are optimized, stripped, and ready for release!"
