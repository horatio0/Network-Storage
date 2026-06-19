#!/bin/bash
set -e

# Change directory to the frontend root (one level up from scripts directory)
cd "$(dirname "$0")/.."

echo "==> Preparing Fyne CLI..."
mkdir -p bin
GOBIN="$(pwd)/bin" go install fyne.io/tools/cmd/fyne@latest

echo "==> Packaging for Linux..."
./bin/fyne package -os linux -icon resources/Icon.jpg

echo "==> Packaging for Windows..."
if command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then
    CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ ./bin/fyne package -os windows -icon resources/Icon.jpg
else
    echo "❌ Error: MinGW-w64 cross-compiler is not installed."
    echo "To build for Windows on Linux, please install it first:"
    echo "  sudo dnf install mingw64-gcc mingw64-gcc-c++"
    exit 1
fi

echo "==> Build complete!"
