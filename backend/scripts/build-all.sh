#!/bin/bash

# Cross-platform build script for Ledabeer Backend
# This script builds the backend for all supported platforms

set -e

echo "🚀 Building Ledabeer Backend for all platforms..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local ext=$3
    local output_name=$4
    
    echo "📦 Building for $os $arch..."
    
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build \
        -ldflags='-w -s -extldflags "-static"' \
        -a -installsuffix cgo \
        -o "bin/ledabeer-$output_name$ext" \
        ./cmd/node
    
    echo "✅ Built: bin/ledabeer-$output_name$ext"
}

# Build for all platforms
echo "🔨 Building Linux binaries..."
build_platform "linux" "amd64" "" "linux-amd64"
build_platform "linux" "arm64" "" "linux-arm64"
build_platform "linux" "arm" "" "linux-arm"

echo "🔨 Building Windows binaries..."
build_platform "windows" "amd64" ".exe" "windows-amd64"
build_platform "windows" "arm64" ".exe" "windows-arm64"

echo "🔨 Building macOS binaries..."
build_platform "darwin" "amd64" "" "darwin-amd64"
build_platform "darwin" "arm64" "" "darwin-arm64"

echo "🔨 Building FreeBSD binaries..."
build_platform "freebsd" "amd64" "" "freebsd-amd64"
build_platform "freebsd" "arm64" "" "freebsd-arm64"

echo "🔨 Building OpenBSD binaries..."
build_platform "openbsd" "amd64" "" "openbsd-amd64"
build_platform "openbsd" "arm64" "" "openbsd-arm64"

echo "🔨 Building NetBSD binaries..."
build_platform "netbsd" "amd64" "" "netbsd-amd64"
build_platform "netbsd" "arm64" "" "netbsd-arm64"

echo "🔨 Building Solaris binaries..."
build_platform "solaris" "amd64" "" "solaris-amd64"

echo ""
echo "🎉 Build complete! All binaries are available in the bin/ directory:"
echo ""
ls -la bin/
echo ""
echo "📋 Platform support:"
echo "  ✅ Linux (AMD64, ARM64, ARM)"
echo "  ✅ Windows (AMD64, ARM64)"
echo "  ✅ macOS (Intel, Apple Silicon)"
echo "  ✅ FreeBSD (AMD64, ARM64)"
echo "  ✅ OpenBSD (AMD64, ARM64)"
echo "  ✅ NetBSD (AMD64, ARM64)"
echo "  ✅ Solaris (AMD64)"
echo ""
echo "🐳 Docker images can be built with:"
echo "  docker build -t ledabeer-backend ."
echo "  docker build -f Dockerfile.windows -t ledabeer-backend-windows ."
echo ""
echo "🚀 Ready for deployment on any platform!"
