# Backend Deployment Guide

## Overview

This guide explains how to deploy the Workflow backend to EC2 using GitHub Actions.

## Prerequisites

- EC2 instance running Ubuntu
- SSH access to EC2
- GitHub repository secrets configured

## Required GitHub Secrets

Configure these secrets in your GitHub repository (Settings → Secrets and variables → Actions):

### For Cloud Environment:
- `EC2_CLOUD_HOST`: EC2 public IP or hostname (e.g., `13.203.37.93`)
- `EC2_CLOUD_USER`: SSH username (e.g., `ubuntu`)
- `EC2_SSH_KEY`: Private SSH key for EC2 access

### For Local Environment:
- `EC2_LOCAL_HOST`: Local EC2 IP or hostname
- `EC2_LOCAL_USER`: SSH username
- `EC2_SSH_KEY`: Private SSH key (same as cloud)

## Deployment Steps

### 1. Initial Server Setup (One-time)

SSH into your EC2 instance and run:

```bash
# Upload setup files to EC2
scp backend/deploy/setup-server.sh ubuntu@13.203.37.93:/tmp/
scp backend/deploy/workflow-backend.service ubuntu@13.203.37.93:/tmp/

# SSH into EC2
ssh ubuntu@13.203.37.93

# Run setup script
cd /tmp
chmod +x setup-server.sh
./setup-server.sh

# Create .env file
sudo nano /opt/workflow/.env
```

Copy the following template and fill in your values:

```bash
# Server Configuration
PORT=7000
GIN_MODE=release

# RabbitMQ Configuration
RABBITMQ_URL=amqp://board:examplepassword@13.203.37.93:5672/
RABBITMQ_QUEUE_NAME=workflow_queue
RABBITMQ_USERNAME=board
RABBITMQ_PASSWORD=examplepassword
RABBITMQ_HOST=13.203.37.93
RABBITMQ_PORT=5672

# Database Configuration
DB_HOST=13.203.37.93
DB_PORT=3306
DB_NAME=dev_workflow
DB_USERNAME=root
DB_PASSWORD=examplepassword

# Environment
ENVIRONMENT=production
LOG_LEVEL=info

# Working Directory
WORKING_DIR=/tmp/claude-tasks

# CORS Configuration
CORS_ORIGIN=http://jokertrickster-workflow.s3-website.ap-south-1.amazonaws.com

# GitHub Configuration
GITHUB_TOKEN=your_github_token_here
```

### 2. Deploy via GitHub Actions

1. Go to GitHub repository → Actions tab
2. Select "Deploy Backend" workflow
3. Click "Run workflow"
4. Choose options:
   - **Branch**: `main` (or your target branch)
   - **Component**: `backend`
   - **Environment**: `cloud` or `local`
5. Click "Run workflow"

### 3. Verify Deployment

SSH into EC2 and check:

```bash
# Check service status
sudo systemctl status workflow-backend

# View logs
sudo journalctl -u workflow-backend -f

# Check if server is running
curl http://localhost:7000/health
```

## Manual Deployment (Alternative)

If GitHub Actions is not available:

```bash
# 1. Build locally
cd backend
go build -o server main.go

# 2. Upload to EC2
scp server ubuntu@13.203.37.93:/tmp/

# 3. Deploy on EC2
ssh ubuntu@13.203.37.93
sudo systemctl stop workflow-backend
sudo cp /tmp/server /opt/workflow/
sudo chmod +x /opt/workflow/server
sudo systemctl start workflow-backend
sudo systemctl status workflow-backend
```

## Troubleshooting

### Service won't start

```bash
# Check logs
sudo journalctl -u workflow-backend -n 50 --no-pager

# Check permissions
ls -la /opt/workflow/
sudo chmod +x /opt/workflow/server

# Check .env file
cat /opt/workflow/.env
```

### Connection issues

```bash
# Check if port 7000 is open
sudo netstat -tulpn | grep 7000

# Check firewall
sudo ufw status

# Test local connection
curl http://localhost:7000/health
```

### Database connection issues

```bash
# Test MySQL connection from EC2
mysql -h 13.203.37.93 -u root -p dev_workflow

# Check environment variables
sudo systemctl show workflow-backend | grep Environment
```

## Rollback

If deployment fails, rollback to previous version:

```bash
ssh ubuntu@13.203.37.93

# List backups
ls -la /opt/workflow/*.backup.*

# Restore backup (replace timestamp with actual backup)
sudo systemctl stop workflow-backend
sudo cp /opt/workflow/server.backup.20250102-103000 /opt/workflow/server
sudo systemctl start workflow-backend
```

## Service Management Commands

```bash
# Start service
sudo systemctl start workflow-backend

# Stop service
sudo systemctl stop workflow-backend

# Restart service
sudo systemctl restart workflow-backend

# Check status
sudo systemctl status workflow-backend

# View logs (live)
sudo journalctl -u workflow-backend -f

# View recent logs
sudo journalctl -u workflow-backend -n 100

# Enable auto-start on boot
sudo systemctl enable workflow-backend

# Disable auto-start
sudo systemctl disable workflow-backend
```

## Security Considerations

1. **SSH Key**: Never commit SSH keys to the repository
2. **Secrets**: Use GitHub Secrets for sensitive data
3. **Environment Variables**: Store in `/opt/workflow/.env` on EC2
4. **Firewall**: Only allow necessary ports (7000 for API)
5. **CORS**: Update `CORS_ORIGIN` to match your S3 website URL

## Monitoring

### Health Check Endpoint

```bash
curl http://13.203.37.93:7000/health
```

### Log Files

- Application logs: `/var/log/workflow-backend/server.log`
- Error logs: `/var/log/workflow-backend/error.log`
- System logs: `sudo journalctl -u workflow-backend`

## Architecture

```
GitHub Actions
    ↓
    1. Build Go binary
    ↓
    2. Upload to EC2 via SCP
    ↓
    3. Stop service
    ↓
    4. Backup old binary
    ↓
    5. Deploy new binary
    ↓
    6. Start service
    ↓
    7. Verify status
```

## Support

For issues or questions:
1. Check logs: `sudo journalctl -u workflow-backend -f`
2. Review this README
3. Create GitHub issue with logs and error details
