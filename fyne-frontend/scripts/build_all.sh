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
/home/gyumin/go/bin/fyne package -os linux -icon ./resources/Icon.png
mv -f *.tar.xz release/

# 2. Build for Windows
echo "----------------------------------------"
echo "[2/3] Packaging for Windows..."
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "Error: x86_64-w64-mingw32-gcc not found."
    echo "Please install the mingw-w64 package to build for Windows."
    exit 1
fi
CC=x86_64-w64-mingw32-gcc /home/gyumin/go/bin/fyne package -os windows -icon ./resources/Icon.png
mv -f *.exe release/

# 3. Build for Android
echo "----------------------------------------"
echo "[3/3] Packaging for Android..."
if [ -z "$ANDROID_HOME" ] && [ -z "$ANDROID_NDK_HOME" ]; then
    echo "Android SDK/NDK가 설정되지 않아 Android 빌드를 건너뜁니다."
else
    /home/gyumin/go/bin/fyne package -os android -appID com.network.storage.client -icon ./resources/Icon.png
    mv -f *.apk release/
fi

echo "----------------------------------------"
echo "Builds completed successfully!"
echo "Release artifacts are located in the fyne-frontend/release/ directory."
