#!/bin/bash
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

cd "$(dirname "$0")/.."

echo "Building Fyne app for Linux..."
/home/gyumin/go/bin/fyne package -os linux -icon ./resources/Icon.png

if [ ! -f *.tar.xz ]; then
  echo "Build failed. No .tar.xz file found."
  exit 1
fi

echo "Installing to /..."
tar -xvf *.tar.xz -C / --strip-components=1

echo "Installation complete."
