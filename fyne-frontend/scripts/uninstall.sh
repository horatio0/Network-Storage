#!/bin/bash

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (or use sudo)"
  exit 1
fi

APP_NAME="network-storage-client"

echo "Uninstalling $APP_NAME..."

echo "Removing binary..."
rm -f "/usr/local/bin/$APP_NAME"

echo "Removing desktop entry..."
rm -f "/usr/local/share/applications/$APP_NAME.desktop"

echo "Removing icon files..."
find /usr/local/share/icons/hicolor -type f -name "$APP_NAME.png" -exec rm -f {} +

echo "Updating icon cache..."
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t /usr/local/share/icons/hicolor || true
else
  echo "gtk-update-icon-cache not found, skipping."
fi

echo "Updating desktop database..."
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database /usr/local/share/applications || true
else
  echo "update-desktop-database not found, skipping."
fi

echo "Uninstallation of $APP_NAME is complete."
