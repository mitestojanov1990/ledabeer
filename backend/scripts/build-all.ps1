# Cross-platform build script for Ledabeer Backend (PowerShell)
# This script builds the backend for all supported platforms

param(
    [switch]$Verbose = $false
)

Write-Host "🚀 Building Ledabeer Backend for all platforms..." -ForegroundColor Green

# Create bin directory if it doesn't exist
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Function to build for a specific platform
function Build-Platform {
    param(
        [string]$OS,
        [string]$Arch,
        [string]$Ext,
        [string]$OutputName
    )
    
    Write-Host "📦 Building for $OS $Arch..." -ForegroundColor Yellow
    
    $env:GOOS = $OS
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"
    
    go build `
        -ldflags='-w -s -extldflags "-static"' `
        -a -installsuffix cgo `
        -o "bin/ledabeer-$OutputName$Ext" `
        ./cmd/node
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Built: bin/ledabeer-$OutputName$Ext" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to build for $OS $Arch" -ForegroundColor Red
        exit 1
    }
}

# Build for all platforms
Write-Host "🔨 Building Linux binaries..." -ForegroundColor Cyan
Build-Platform "linux" "amd64" "" "linux-amd64"
Build-Platform "linux" "arm64" "" "linux-arm64"
Build-Platform "linux" "arm" "" "linux-arm"

Write-Host "🔨 Building Windows binaries..." -ForegroundColor Cyan
Build-Platform "windows" "amd64" ".exe" "windows-amd64"
Build-Platform "windows" "arm64" ".exe" "windows-arm64"

Write-Host "🔨 Building macOS binaries..." -ForegroundColor Cyan
Build-Platform "darwin" "amd64" "" "darwin-amd64"
Build-Platform "darwin" "arm64" "" "darwin-arm64"

Write-Host "🔨 Building FreeBSD binaries..." -ForegroundColor Cyan
Build-Platform "freebsd" "amd64" "" "freebsd-amd64"
Build-Platform "freebsd" "arm64" "" "freebsd-arm64"

Write-Host "🔨 Building OpenBSD binaries..." -ForegroundColor Cyan
Build-Platform "openbsd" "amd64" "" "openbsd-amd64"
Build-Platform "openbsd" "arm64" "" "openbsd-arm64"

Write-Host "🔨 Building NetBSD binaries..." -ForegroundColor Cyan
Build-Platform "netbsd" "amd64" "" "netbsd-amd64"
Build-Platform "netbsd" "arm64" "" "netbsd-arm64"

Write-Host "🔨 Building Solaris binaries..." -ForegroundColor Cyan
Build-Platform "solaris" "amd64" "" "solaris-amd64"

Write-Host ""
Write-Host "🎉 Build complete! All binaries are available in the bin/ directory:" -ForegroundColor Green
Write-Host ""
Get-ChildItem -Path "bin" | Format-Table Name, Length, LastWriteTime
Write-Host ""
Write-Host "📋 Platform support:" -ForegroundColor Blue
Write-Host "  ✅ Linux (AMD64, ARM64, ARM)"
Write-Host "  ✅ Windows (AMD64, ARM64)"
Write-Host "  ✅ macOS (Intel, Apple Silicon)"
Write-Host "  ✅ FreeBSD (AMD64, ARM64)"
Write-Host "  ✅ OpenBSD (AMD64, ARM64)"
Write-Host "  ✅ NetBSD (AMD64, ARM64)"
Write-Host "  ✅ Solaris (AMD64)"
Write-Host ""
Write-Host "🐳 Docker images can be built with:" -ForegroundColor Magenta
Write-Host "  docker build -t ledabeer-backend ."
Write-Host "  docker build -f Dockerfile.windows -t ledabeer-backend-windows ."
Write-Host ""
Write-Host "🚀 Ready for deployment on any platform!" -ForegroundColor Green
