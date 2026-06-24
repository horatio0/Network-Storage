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

echo "[1/4] Building backend binary for ARM64 (Linux) using go build..."
# Build the binary for ARM architecture (Raspberry Pi, etc.)
env GOOS=linux GOARCH=arm64 go build -o NS-server .

echo "[2/4] Moving binary to /NS/server/NS-server..."
mkdir -p /NS/server
mv NS-server /NS/server/NS-server
chmod +x /NS/server/NS-server

if [ -f "/etc/systemd/system/NS.service" ]; then
  echo "Service file already exists. Skipping creation."
  echo "Restarting the service..."
  systemctl restart NS.service
else
  echo "[3/4] Creating systemd service file..."
  cat > /etc/systemd/system/NS.service << 'EOF'
[Unit]
Description=Network Storage Backend Server
After=network.target

[Service]
Type=simple
# Change User to a non-root user if needed
User=hortio
ExecStart=/NS/server/NS-server
Restart=on-failure
RestartSec=5
# Adjust WorkingDirectory if your server requires specific paths
WorkingDirectory=/NS/server

[Install]
WantedBy=multi-user.target
EOF

  echo "[4/4] Reloading systemd and enabling the service..."
  systemctl daemon-reload
  systemctl enable --now NS.service
fi

echo "Installation complete! The backend server is now up-to-date and running."
echo "Check status: systemctl status NS.service"
