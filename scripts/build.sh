#!/bin/bash

set -e

# Get current directory
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$DIR/.." && pwd)"

# Set environment variables
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# Create build directory if it doesn't exist
if [ ! -d "$PROJECT_ROOT/bin" ]; then
    mkdir -p "$PROJECT_ROOT/bin"
fi

# Clean previous build
rm -f "$PROJECT_ROOT/bin/bookshop-api"

echo "==> Downloading dependencies..."
cd "$PROJECT_ROOT" && go mod download

echo "==> Building application..."
cd "$PROJECT_ROOT" && go build -o "$PROJECT_ROOT/bin/bookshop-api" -ldflags="-s -w" ./cmd/api

echo "==> Build completed successfully!"
echo "Executable: $PROJECT_ROOT/bin/bookshop-api"


