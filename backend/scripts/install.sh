#!/bin/bash
# Network Storage Backend Installation Script for Ubuntu
# 
# Usage instructions:
# 1. Grant execution permission: chmod +x install.sh
# 2. Run as root: sudo ./install.sh

set -e

# Require root privilege
if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run this script as root (e.g. sudo ./install.sh)"
  exit 1
fi

# Navigate to backend directory
cd "$(dirname "$0")/.."

echo "[1/4] Building backend binary using go build..."
# Build the binary
go build -o network-storage-server .

echo "[2/4] Moving binary to /NS/server/network-storage-server..."
mkdir -p /NS/server
mv network-storage-server /NS/server/network-storage-server
chmod +x /NS/server/network-storage-server

echo "[3/4] Creating systemd service file..."
cat > /etc/systemd/system/network-storage.service << 'EOF'
[Unit]
Description=Network Storage Backend Server
After=network.target

[Service]
Type=simple
# Change User to a non-root user if needed
User=root
ExecStart=/NS/server/network-storage-server
Restart=on-failure
RestartSec=5
# Adjust WorkingDirectory if your server requires specific paths
WorkingDirectory=/NS/server

[Install]
WantedBy=multi-user.target
EOF

echo "[4/4] Reloading systemd and enabling the service..."
systemctl daemon-reload
systemctl enable --now network-storage

echo "Installation complete! The backend server is now running."
echo "Check status: systemctl status network-storage"
