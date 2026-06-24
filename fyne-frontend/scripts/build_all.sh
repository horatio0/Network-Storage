#!/bin/bash

# ==============================================================================
# Fyne Frontend Build Script
# 
# Prerequisites:
# 1. Fyne CLI tool must be installed: `go install fyne.io/fyne/v2/cmd/fyne@latest`
# 2. Android SDK and NDK must be installed and configured in your environment 
#    for the Android build (e.g., ANDROID_HOME, ANDROID_NDK_HOME).
# 3. C compiler (gcc/clang) for Linux build.
# 4. MinGW-w64 (e.g., x86_64-w64-mingw32-gcc) for Windows cross-compilation from Linux.
# ==============================================================================

set -e

echo "Starting Fyne client builds for Linux, Windows, and Android..."

# Navigate to the fyne-frontend project root (one directory up from scripts)
cd "$(dirname "$0")/.."

# Create the release directory if it doesn't exist
mkdir -p release

# 1. Build for Linux
echo "----------------------------------------"
echo "[1/3] Packaging for Linux..."
/home/gyumin/go/bin/fyne package -os linux
mv -f *.tar.xz release/

# 2. Build for Windows
echo "----------------------------------------"
echo "[2/3] Packaging for Windows..."
/home/gyumin/go/bin/fyne package -os windows
mv -f *.exe release/

# 3. Build for Android
echo "----------------------------------------"
echo "[3/3] Packaging for Android..."
/home/gyumin/go/bin/fyne package -os android -appID com.network.storage.client
mv -f *.apk release/

echo "----------------------------------------"
echo "Builds completed successfully!"
echo "Release artifacts are located in the fyne-frontend/release/ directory."
