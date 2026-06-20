#!/bin/bash
set -e

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (e.g., sudo ./scripts/uninstall.sh)"
  exit 1
fi

echo "Uninstalling network-storage-client..."

rm -f /usr/local/bin/network-storage-client
rm -f /usr/local/share/applications/network-storage-client.desktop
rm -f /usr/local/share/pixmaps/network-storage-client.png

echo "Uninstallation complete."
