#!/bin/bash

# Default values
PORT=${PORT:-8080}
ENVIRONMENT=${ENVIRONMENT:-development}
BUILD_DIR="./bin"
APP_NAME="bookstore"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Navigate to project root
cd "$(dirname "$0")/../../" || exit

# Create build directory if it doesn't exist
mkdir -p "$BUILD_DIR"

echo "📦 Building $APP_NAME (env: $ENVIRONMENT, port: $PORT)"
echo "----------------------------------------"

# Clean
echo "🧹 Cleaning previous builds..."
go clean

# Download dependencies
echo "📥 Downloading dependencies..."
go mod download

# Build for current platform
echo "🔨 Building application..."
go build -o "$BUILD_DIR/$APP_NAME" ./cmd/main/main.go

# Check if build succeeded
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

# Make the binary executable
chmod +x "$BUILD_DIR/$APP_NAME"

echo "🚀 Starting $APP_NAME on port $PORT..."
export PORT=$PORT
export ENVIRONMENT=$ENVIRONMENT
"$BUILD_DIR/$APP_NAME"