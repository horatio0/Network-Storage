#!/bin/bash
set -e

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (e.g., sudo ./scripts/install.sh)"
  exit 1
fi

TAR_FILE="$(dirname "$0")/../network-storage-client.tar.xz"

if [ ! -f "$TAR_FILE" ]; then
  ./build.sh
fi

echo "Installing network-storage-client to / ..."
tar -xJf "$TAR_FILE" -C / --strip-components=1

echo "Installation complete."
