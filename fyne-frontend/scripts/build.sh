#!/bin/bash
set -e

echo "==> Packaging for Windows..."
fyne package -os windows -icon resources/Icon.jpg

echo "==> Packaging for Linux..."
fyne package -os linux -icon resources/Icon.jpg

echo "==> Build complete!"
