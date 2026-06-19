#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Change directory to the backend root (one level up from scripts directory)
cd "$(dirname "$0")/.."

echo "Building backend..."

# Create a bin directory for the output if it doesn't exist
mkdir -p bin

# Build the Go application into a single executable binary
# go build automatically compiles the main package and all its imported internal packages into one executable file.
go build -o bin/backend-server main.go

echo "Build successful!"
echo "Executable is located at: bin/backend-server"
