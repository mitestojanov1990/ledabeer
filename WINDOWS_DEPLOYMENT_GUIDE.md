# Ledabeer Backend - Windows Deployment Guide

## 🪟 Windows Deployment Options

The Ledabeer backend can be deployed on Windows in multiple ways:

1. **Native Windows Binary** - Direct execution
2. **Docker Desktop** - Containerized deployment
3. **Windows Service** - Production service
4. **WSL2** - Linux subsystem

## 🚀 Quick Start

### Option 1: Native Windows Binary

#### Prerequisites
- **Go 1.21+**: Download from https://golang.org/dl/
- **Git**: Download from https://git-scm.com/download/win

#### Installation
```powershell
# Clone repository
git clone <repository-url>
cd ledabeer\backend

# Build for Windows
go build -o bin\ledabeer.exe .\cmd\node

# Run the backend
.\bin\ledabeer.exe
```

### Option 2: Docker Desktop (Recommended)

#### Prerequisites
- **Docker Desktop**: Download from https://www.docker.com/products/docker-desktop/
- **WSL2**: Enable WSL2 backend in Docker Desktop

#### Installation
```powershell
# Clone repository
git clone <repository-url>
cd ledabeer\backend

# Build and run with Docker
docker build -t ledabeer-backend .
docker run -p 50051:50051 -p 8080:8080 -p 4001:4001 ledabeer-backend
```

#### Docker Compose
```powershell
# Start multi-node cluster
docker-compose -f docker-compose.cross-platform.yml up -d

# Check status
docker-compose ps
docker-compose logs -f
```

## 🔧 Advanced Windows Deployment

### Windows Service Installation

#### Using NSSM (Non-Sucking Service Manager)

1. **Download NSSM**: https://nssm.cc/download
2. **Extract NSSM**: Extract to `C:\nssm\`
3. **Install Service**:
```powershell
# Install the service
C:\nssm\win64\nssm.exe install LedabeerBackend

# Configure the service
C:\nssm\win64\nssm.exe set LedabeerBackend Application "C:\ledabeer\bin\ledabeer.exe"
C:\nssm\win64\nssm.exe set LedabeerBackend AppDirectory "C:\ledabeer"
C:\nssm\win64\nssm.exe set LedabeerBackend DisplayName "Ledabeer Backend"
C:\nssm\win64\nssm.exe set LedabeerBackend Description "Ledabeer P2P Chat Backend Service"
C:\nssm\win64\nssm.exe set LedabeerBackend Start SERVICE_AUTO_START

# Set environment variables
C:\nssm\win64\nssm.exe set LedabeerBackend AppEnvironmentExtra LOG_LEVEL=info
C:\nssm\win64\nssm.exe set LedabeerBackend AppEnvironmentExtra LOG_FORMAT=json

# Start the service
C:\nssm\win64\nssm.exe start LedabeerBackend
```

#### Using Windows Service Manager

1. **Create Service**:
```powershell
# Create service using sc command
sc create "LedabeerBackend" binPath="C:\ledabeer\bin\ledabeer.exe" start=auto
sc description "LedabeerBackend" "Ledabeer P2P Chat Backend Service"

# Start the service
sc start LedabeerBackend
```

2. **Configure Service**:
```powershell
# Set service to start automatically
sc config LedabeerBackend start=auto

# Set service to restart on failure
sc failure LedabeerBackend reset=86400 actions=restart/5000/restart/10000/restart/20000
```

### Windows Firewall Configuration

```powershell
# Allow inbound connections on required ports
New-NetFirewallRule -DisplayName "Ledabeer gRPC" -Direction Inbound -Protocol TCP -LocalPort 50051 -Action Allow
New-NetFirewallRule -DisplayName "Ledabeer HTTP" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
New-NetFirewallRule -DisplayName "Ledabeer P2P" -Direction Inbound -Protocol TCP -LocalPort 4001 -Action Allow

# Allow outbound connections
New-NetFirewallRule -DisplayName "Ledabeer Outbound" -Direction Outbound -Protocol TCP -Action Allow
```

### Windows Performance Optimization

```powershell
# Optimize network settings
netsh int tcp set global autotuninglevel=normal
netsh int tcp set global chimney=enabled
netsh int tcp set global rss=enabled
netsh int tcp set global netdma=enabled

# Optimize memory settings
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management" /v LargeSystemCache /t REG_DWORD /d 0 /f
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management" /v DisablePagingExecutive /t REG_DWORD /d 1 /f
```

## 🐳 Docker on Windows

### Docker Desktop Configuration

1. **Enable WSL2 Backend**:
   - Open Docker Desktop
   - Go to Settings → General
   - Check "Use the WSL 2 based engine"
   - Apply & Restart

2. **Configure Resources**:
   - Go to Settings → Resources
   - Set CPU: 4+ cores
   - Set Memory: 4GB+
   - Set Disk: 20GB+

### Docker Compose for Windows

```yaml
# docker-compose.windows.yml
version: "3.8"

services:
  ledabeer-backend:
    build:
      context: .
      dockerfile: Dockerfile.windows
    container_name: ledabeer-backend-windows
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
    driver: nat
```

### Running Docker on Windows

```powershell
# Build Windows-specific image
docker build -f Dockerfile.windows -t ledabeer-backend-windows .

# Run with Windows-specific settings
docker run -d `
  --name ledabeer-backend `
  -p 50051:50051 `
  -p 8080:8080 `
  -p 4001:4001 `
  -v ${PWD}\data:/app/data `
  -v ${PWD}\logs:/app/logs `
  ledabeer-backend-windows

# Check status
docker ps
docker logs ledabeer-backend
```

## 🔍 Windows-Specific Troubleshooting

### Common Issues

1. **Port Already in Use**:
```powershell
# Check what's using the port
netstat -ano | findstr :50051
netstat -ano | findstr :8080
netstat -ano | findstr :4001

# Kill the process
taskkill /PID <PID> /F
```

2. **Firewall Blocking**:
```powershell
# Check firewall status
Get-NetFirewallRule -DisplayName "*Ledabeer*"

# Temporarily disable firewall (not recommended for production)
Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled False
```

3. **Antivirus Interference**:
   - Add `C:\ledabeer\` to antivirus exclusions
   - Whitelist the `ledabeer.exe` executable
   - Disable real-time scanning for the application directory

4. **Permission Issues**:
```powershell
# Run PowerShell as Administrator
# Grant full control to the application directory
icacls "C:\ledabeer" /grant "Users:(OI)(CI)F" /T
```

### Performance Monitoring

```powershell
# Monitor CPU and Memory usage
Get-Process -Name "ledabeer" | Select-Object CPU, WorkingSet, ProcessName

# Monitor network connections
netstat -ano | findstr "ledabeer"

# Monitor disk usage
Get-WmiObject -Class Win32_LogicalDisk | Select-Object DeviceID, Size, FreeSpace
```

### Log Management

```powershell
# View Windows Event Log
Get-WinEvent -LogName Application | Where-Object {$_.ProviderName -eq "Ledabeer"}

# View application logs
Get-Content "C:\ledabeer\logs\*.log" -Tail 100

# Rotate logs
Get-ChildItem "C:\ledabeer\logs\*.log" | Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-7)} | Remove-Item
```

## 🚀 Production Deployment

### Windows Server Deployment

1. **Server Requirements**:
   - Windows Server 2019/2022
   - 4+ CPU cores
   - 8GB+ RAM
   - 50GB+ storage
   - Network connectivity

2. **Installation Steps**:
```powershell
# Install as Windows Service
C:\nssm\win64\nssm.exe install LedabeerBackend
C:\nssm\win64\nssm.exe set LedabeerBackend Application "C:\ledabeer\bin\ledabeer.exe"
C:\nssm\win64\nssm.exe set LedabeerBackend AppDirectory "C:\ledabeer"
C:\nssm\win64\nssm.exe set LedabeerBackend Start SERVICE_AUTO_START

# Configure firewall
New-NetFirewallRule -DisplayName "Ledabeer gRPC" -Direction Inbound -Protocol TCP -LocalPort 50051 -Action Allow
New-NetFirewallRule -DisplayName "Ledabeer HTTP" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
New-NetFirewallRule -DisplayName "Ledabeer P2P" -Direction Inbound -Protocol TCP -LocalPort 4001 -Action Allow

# Start service
C:\nssm\win64\nssm.exe start LedabeerBackend
```

3. **Monitoring Setup**:
```powershell
# Install monitoring tools
# Use Windows Performance Monitor
# Set up alerts for CPU, Memory, Disk usage
# Configure log rotation
```

### High Availability Setup

1. **Load Balancer**: Use Windows NLB or external load balancer
2. **Database**: Use SQL Server for persistent storage
3. **Backup**: Implement automated backup strategy
4. **Monitoring**: Use Windows Performance Monitor and Event Log

## 📋 Windows Deployment Checklist

### Pre-Deployment
- [ ] Verify Windows version compatibility
- [ ] Install required dependencies (Go, Git)
- [ ] Configure Windows Firewall
- [ ] Set up antivirus exclusions
- [ ] Plan backup strategy

### Deployment
- [ ] Build Windows binary
- [ ] Configure environment variables
- [ ] Install as Windows Service
- [ ] Configure firewall rules
- [ ] Test connectivity

### Post-Deployment
- [ ] Monitor performance
- [ ] Check Windows Event Log
- [ ] Verify service status
- [ ] Test all endpoints
- [ ] Document configuration

## 🆘 Windows-Specific Support

### Getting Help
1. **Windows Event Log**: Check Application and System logs
2. **Performance Monitor**: Monitor system resources
3. **Network Diagnostics**: Use `netsh` and `netstat` commands
4. **Service Management**: Use `sc` and `services.msc`

### Debug Commands
```powershell
# Check service status
sc query LedabeerBackend
sc queryex LedabeerBackend

# Check network connections
netstat -ano | findstr "ledabeer"

# Check firewall rules
Get-NetFirewallRule -DisplayName "*Ledabeer*"

# Check process information
Get-Process -Name "ledabeer" | Format-List *
```

## 🎯 Conclusion

The Ledabeer backend is fully compatible with Windows and can be deployed using:

- ✅ **Native Windows Binary** - Direct execution
- ✅ **Docker Desktop** - Containerized deployment  
- ✅ **Windows Service** - Production service
- ✅ **WSL2** - Linux subsystem

Choose the deployment method that best fits your environment and requirements! 🚀
