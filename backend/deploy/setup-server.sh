#!/bin/bash

# Workflow Backend Server Setup Script for EC2
# This script sets up the backend server on EC2 instance

set -e

echo "🚀 Setting up Workflow Backend Server..."

# Create application directory
sudo mkdir -p /opt/workflow
sudo mkdir -p /var/log/workflow-backend

# Set permissions
sudo chown -R ubuntu:ubuntu /opt/workflow
sudo chown -R ubuntu:ubuntu /var/log/workflow-backend

# Copy environment file (should be done manually or via secrets)
if [ ! -f "/opt/workflow/.env" ]; then
    echo "⚠️  Warning: /opt/workflow/.env not found"
    echo "Please create it with the following template:"
    cat << 'EOF'
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
EOF
fi

# Copy systemd service file
sudo cp workflow-backend.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable workflow-backend

echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Create /opt/workflow/.env with your configuration"
echo "2. Copy the binary: sudo cp server /opt/workflow/"
echo "3. Start the service: sudo systemctl start workflow-backend"
echo "4. Check status: sudo systemctl status workflow-backend"
echo "5. View logs: sudo journalctl -u workflow-backend -f"
