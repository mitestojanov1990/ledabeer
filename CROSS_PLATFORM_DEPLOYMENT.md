# Ledabeer Backend - Cross-Platform Deployment Guide

## 🎯 Overview

This guide covers deploying the Ledabeer backend on multiple platforms:
- **Linux** (Ubuntu, CentOS, Debian)
- **Windows** (Windows 10/11, Windows Server)
- **macOS** (Intel and Apple Silicon)
- **Docker** (Cross-platform containerization)

## 🐧 Linux Deployment

### Prerequisites
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y golang-go git make build-essential

# CentOS/RHEL
sudo yum install -y golang git make gcc
# or for newer versions:
sudo dnf install -y golang git make gcc
```

### Build and Run
```bash
# Clone repository
git clone <repository-url>
cd ledabeer/backend

# Build for Linux
make build

# Run the backend
./bin/node
```

### Systemd Service (Production)
```bash
# Create service file
sudo tee /etc/systemd/system/ledabeer.service > /dev/null <<EOF
[Unit]
Description=Ledabeer Backend Service
After=network.target

[Service]
Type=simple
User=ledabeer
WorkingDirectory=/opt/ledabeer/backend
ExecStart=/opt/ledabeer/backend/bin/node
Restart=always
RestartSec=5
Environment=LOG_LEVEL=info
Environment=LOG_FORMAT=json

[Install]
WantedBy=multi-user.target
EOF

# Create user and directory
sudo useradd -r -s /bin/false ledabeer
sudo mkdir -p /opt/ledabeer/backend
sudo cp -r . /opt/ledabeer/backend/
sudo chown -R ledabeer:ledabeer /opt/ledabeer/

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable ledabeer
sudo systemctl start ledabeer
sudo systemctl status ledabeer
```

## 🪟 Windows Deployment

### Prerequisites
1. **Go 1.21+**: Download from https://golang.org/dl/
2. **Git**: Download from https://git-scm.com/download/win
3. **Make**: Install via Chocolatey or use PowerShell scripts

### Option 1: Using Chocolatey
```powershell
# Install Chocolatey (if not installed)
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install dependencies
choco install golang git make -y
```

### Option 2: Manual Installation
1. Download and install Go from https://golang.org/dl/
2. Download and install Git from https://git-scm.com/download/win
3. Download Make from https://www.gnu.org/software/make/ or use PowerShell

### Build and Run
```powershell
# Clone repository
git clone <repository-url>
cd ledabeer\backend

# Build for Windows
go build -o bin\node.exe .\cmd\node

# Run the backend
.\bin\node.exe
```

### Windows Service (Production)
```powershell
# Install NSSM (Non-Sucking Service Manager)
# Download from https://nssm.cc/download

# Create service
nssm install LedabeerBackend
nssm set LedabeerBackend Application "C:\ledabeer\backend\bin\node.exe"
nssm set LedabeerBackend AppDirectory "C:\ledabeer\backend"
nssm set LedabeerBackend AppParameters ""
nssm set LedabeerBackend DisplayName "Ledabeer Backend"
nssm set LedabeerBackend Description "Ledabeer P2P Chat Backend Service"

# Start service
nssm start LedabeerBackend
```

## 🍎 macOS Deployment

### Prerequisites
```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install go git make
```

### Build and Run
```bash
# Clone repository
git clone <repository-url>
cd ledabeer/backend

# Build for macOS
make build

# Run the backend
./bin/node
```

### LaunchAgent (Production)
```bash
# Create LaunchAgent plist
cat > ~/Library/LaunchAgents/com.ledabeer.backend.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ledabeer.backend</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/node</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/usr/local/ledabeer/backend</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>EnvironmentVariables</key>
    <dict>
        <key>LOG_LEVEL</key>
        <string>info</string>
        <key>LOG_FORMAT</key>
        <string>json</string>
    </dict>
</dict>
</plist>
EOF

# Load the service
launchctl load ~/Library/LaunchAgents/com.ledabeer.backend.plist
launchctl start com.ledabeer.backend
```

## 🐳 Docker Deployment

### Cross-Platform Docker Support

**Yes, Docker will run on Windows!** Docker Desktop for Windows provides full Linux container support.

### Prerequisites
- **Docker Desktop**: https://www.docker.com/products/docker-desktop/
- **Docker Compose**: Included with Docker Desktop

### Single Node Deployment
```bash
# Build and run single node
cd ledabeer/backend
docker build -t ledabeer-backend .
docker run -p 50051:50051 -p 8080:8080 -p 4001:4001 ledabeer-backend
```

### Multi-Node Cluster
```bash
# Start multi-node cluster
cd ledabeer/backend
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### Production Docker Setup
```yaml
# docker-compose.prod.yml
version: "3.8"

services:
  ledabeer-backend:
    build: .
    container_name: ledabeer-backend-prod
    restart: unless-stopped
    ports:
      - "50051:50051"  # gRPC
      - "8080:8080"    # HTTP/WebSocket
      - "4001:4001"    # Libp2p
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      - NODE_ROLE=primary
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    networks:
      - ledabeer-network

networks:
  ledabeer-network:
    driver: bridge
```

## 🔧 Cross-Platform Build Scripts

### Build Script for All Platforms
```bash
#!/bin/bash
# build-all.sh

echo "Building Ledabeer Backend for all platforms..."

# Linux AMD64
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o bin/ledabeer-linux-amd64 ./cmd/node

# Linux ARM64
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o bin/ledabeer-linux-arm64 ./cmd/node

# Windows AMD64
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -o bin/ledabeer-windows-amd64.exe ./cmd/node

# Windows ARM64
echo "Building for Windows ARM64..."
GOOS=windows GOARCH=arm64 go build -o bin/ledabeer-windows-arm64.exe ./cmd/node

# macOS AMD64 (Intel)
echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -o bin/ledabeer-darwin-amd64 ./cmd/node

# macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -o bin/ledabeer-darwin-arm64 ./cmd/node

echo "Build complete! Binaries available in bin/ directory"
```

### PowerShell Build Script (Windows)
```powershell
# build-all.ps1

Write-Host "Building Ledabeer Backend for all platforms..."

# Linux AMD64
Write-Host "Building for Linux AMD64..."
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/ledabeer-linux-amd64 ./cmd/node

# Linux ARM64
Write-Host "Building for Linux ARM64..."
$env:GOOS="linux"; $env:GOARCH="arm64"; go build -o bin/ledabeer-linux-arm64 ./cmd/node

# Windows AMD64
Write-Host "Building for Windows AMD64..."
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/ledabeer-windows-amd64.exe ./cmd/node

# Windows ARM64
Write-Host "Building for Windows ARM64..."
$env:GOOS="windows"; $env:GOARCH="arm64"; go build -o bin/ledabeer-windows-arm64.exe ./cmd/node

# macOS AMD64
Write-Host "Building for macOS AMD64..."
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o bin/ledabeer-darwin-amd64 ./cmd/node

# macOS ARM64
Write-Host "Building for macOS ARM64..."
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o bin/ledabeer-darwin-arm64 ./cmd/node

Write-Host "Build complete! Binaries available in bin/ directory"
```

## 🚀 Deployment Strategies

### 1. Development Deployment
```bash
# Quick start for development
git clone <repository>
cd ledabeer/backend
make build
./bin/node
```

### 2. Production Deployment
```bash
# Production deployment with Docker
git clone <repository>
cd ledabeer/backend
docker-compose -f docker-compose.prod.yml up -d
```

### 3. Cloud Deployment
```bash
# Deploy to cloud provider
docker build -t ledabeer-backend .
docker tag ledabeer-backend your-registry/ledabeer-backend:latest
docker push your-registry/ledabeer-backend:latest

# Deploy to cloud
kubectl apply -f k8s/
```

## 🔍 Platform-Specific Considerations

### Linux
- **Systemd**: Use systemd for service management
- **Firewall**: Configure iptables/ufw for port access
- **Permissions**: Run as non-root user
- **Logs**: Use journald for logging

### Windows
- **Windows Service**: Use NSSM or Windows Service Manager
- **Firewall**: Configure Windows Firewall
- **Antivirus**: Whitelist the application
- **Logs**: Use Windows Event Log

### macOS
- **LaunchAgent**: Use LaunchAgent for service management
- **Firewall**: Configure macOS Firewall
- **Permissions**: Grant necessary permissions
- **Logs**: Use Console.app for logging

### Docker
- **Cross-platform**: Works on all platforms
- **Isolation**: Provides process isolation
- **Scaling**: Easy horizontal scaling
- **Orchestration**: Use Docker Swarm or Kubernetes

## 📊 Performance Considerations

### Resource Requirements
- **CPU**: 1-2 cores minimum, 4+ cores recommended
- **RAM**: 512MB minimum, 2GB+ recommended
- **Storage**: 1GB minimum, 10GB+ recommended
- **Network**: 100Mbps minimum, 1Gbps+ recommended

### Optimization
```bash
# Linux: Optimize for performance
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
sysctl -p

# Windows: Optimize network settings
netsh int tcp set global autotuninglevel=normal
netsh int tcp set global chimney=enabled
```

## 🔒 Security Considerations

### Network Security
- **Firewall**: Configure appropriate firewall rules
- **TLS**: Use TLS for gRPC connections
- **VPN**: Use VPN for secure connections
- **NAT**: Configure NAT traversal

### Application Security
- **User Permissions**: Run with minimal privileges
- **File Permissions**: Restrict file access
- **Logging**: Enable security logging
- **Updates**: Keep dependencies updated

## 🧪 Testing Cross-Platform

### Automated Testing
```bash
# Test on multiple platforms
docker run --rm -v $(pwd):/app -w /app golang:1.21 go test ./...

# Test Docker build
docker build -t ledabeer-test .
docker run --rm ledabeer-test
```

### Manual Testing
1. **Linux**: Test on Ubuntu, CentOS, Debian
2. **Windows**: Test on Windows 10/11, Windows Server
3. **macOS**: Test on Intel and Apple Silicon
4. **Docker**: Test on all platforms

## 📋 Deployment Checklist

### Pre-Deployment
- [ ] Verify Go version compatibility
- [ ] Check system requirements
- [ ] Configure firewall rules
- [ ] Set up monitoring
- [ ] Plan backup strategy

### Deployment
- [ ] Build application
- [ ] Configure environment
- [ ] Start services
- [ ] Verify connectivity
- [ ] Test functionality

### Post-Deployment
- [ ] Monitor performance
- [ ] Check logs
- [ ] Verify security
- [ ] Update documentation
- [ ] Plan maintenance

## 🆘 Troubleshooting

### Common Issues
1. **Port Conflicts**: Check if ports are already in use
2. **Permission Issues**: Ensure proper user permissions
3. **Network Issues**: Verify firewall and network configuration
4. **Dependency Issues**: Check Go version and dependencies

### Debug Commands
```bash
# Check if ports are in use
netstat -tulpn | grep :50051
netstat -tulpn | grep :8080

# Check service status
systemctl status ledabeer
docker-compose ps

# Check logs
journalctl -u ledabeer -f
docker-compose logs -f
```

## 📞 Support

For platform-specific issues:
- **Linux**: Check systemd logs and firewall configuration
- **Windows**: Check Windows Event Log and firewall settings
- **macOS**: Check Console.app and system preferences
- **Docker**: Check Docker logs and container status

The Ledabeer backend is designed to be truly cross-platform and will run on any system that supports Go applications or Docker containers! 🚀
