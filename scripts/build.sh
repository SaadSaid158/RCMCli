#!/bin/bash
# RCMCli Build Script - Simple & Clean

set -e

VERSION="1.0.0"
DIST_DIR="dist"

echo "[*] RCMCli Build Script v$VERSION"
echo ""

# Initialize dependencies if needed
if [ ! -f "go.sum" ]; then
    echo "[*] Initializing dependencies..."
    go mod download
    go mod tidy
    echo "[+] Dependencies ready"
    echo ""
fi

# Create dist directories
mkdir -p "$DIST_DIR/Linux"
mkdir -p "$DIST_DIR/Windows"

# Build for all platforms
echo "[*] Building for all platforms..."
echo ""

# Linux x86_64 (native, with CGO)
echo "[*] Building for Linux x86_64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o "$DIST_DIR/Linux/RCMCli_${VERSION}_linux_amd64"
echo "[+] Built: $DIST_DIR/Linux/RCMCli_${VERSION}_linux_amd64"

# Linux ARM64 (cross-compile without CGO - pure Go)
echo "[*] Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o "$DIST_DIR/Linux/RCMCli_${VERSION}_linux_arm64"
if [ $? -eq 0 ]; then
    echo "[+] Built: $DIST_DIR/Linux/RCMCli_${VERSION}_linux_arm64"
else
    echo "[-] Build failed for Linux ARM64"
fi

# Windows x86_64 (cross-compile without CGO - pure Go)
echo "[*] Building for Windows x86_64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o "$DIST_DIR/Windows/RCMCli_${VERSION}_windows_amd64.exe"
if [ $? -eq 0 ]; then
    echo "[+] Built: $DIST_DIR/Windows/RCMCli_${VERSION}_windows_amd64.exe"
else
    echo "[-] Build failed for Windows x86_64"
fi

# Windows ARM64 (cross-compile without CGO - pure Go)
echo "[*] Building for Windows ARM64..."
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o "$DIST_DIR/Windows/RCMCli_${VERSION}_windows_arm64.exe"
if [ $? -eq 0 ]; then
    echo "[+] Built: $DIST_DIR/Windows/RCMCli_${VERSION}_windows_arm64.exe"
else
    echo "[-] Build failed for Windows ARM64"
fi

echo ""
echo "[*] Generating checksums..."
cd "$DIST_DIR"
sha256sum Linux/* Windows/* > checksums.txt
cd ..

echo "[+] All builds complete!"
echo ""
echo "Build artifacts:"
find "$DIST_DIR/" -type f | sort
echo ""
echo "Checksums:"
cat "$DIST_DIR/checksums.txt"
