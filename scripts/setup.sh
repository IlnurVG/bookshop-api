#!/bin/bash

# Set execution permissions
chmod +x scripts/setup.sh
chmod +x scripts/build.sh

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker to run dependencies."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Docker Compose is not installed. Please install Docker Compose to run dependencies."
    exit 1
fi

# Check and install Go dependencies
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

if ! command -v mockgen &> /dev/null; then
    echo "Installing mockgen..."
    go install github.com/golang/mock/mockgen@latest
fi

if ! command -v migrate &> /dev/null; then
    echo "Installing migrate..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Create build directory
mkdir -p bin

echo "==> Creating go.sum..."
go mod tidy

echo "==> Environment setup completed."
echo "To initialize the project, run: make init"
echo "To start the application, run: make run"


