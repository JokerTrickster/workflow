#!/bin/bash

# EC2 서비스 설정 스크립트
# 이 스크립트는 EC2 인스턴스에서 실행되어 systemd 서비스를 설정합니다.

set -e

echo "Setting up workflow services on EC2..."

# 디렉토리 생성
sudo mkdir -p /opt/workflow
sudo mkdir -p /etc/systemd/system

# Backend 서비스 설정
cat <<EOF | sudo tee /etc/systemd/system/workflow-backend.service
[Unit]
Description=Workflow Backend Service
After=network.target

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=/opt/workflow
ExecStart=/opt/workflow/server
Restart=always
RestartSec=5
Environment=PORT=8082
Environment=DB_HOST=localhost
Environment=DB_PORT=3306
Environment=DB_NAME=workflow
Environment=DB_USERNAME=workflow_user
Environment=DB_PASSWORD=workflow_password
Environment=RABBITMQ_URL=amqp://board:examplepassword@13.203.37.93:5672/
Environment=RABBITMQ_QUEUE_NAME=workflow_queue

[Install]
WantedBy=multi-user.target
EOF

# Local Backend 서비스 설정
cat <<EOF | sudo tee /etc/systemd/system/workflow-local-backend.service
[Unit]
Description=Workflow Local Backend Service
After=network.target

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=/opt/workflow
ExecStart=/opt/workflow/local-server
Restart=always
RestartSec=5
Environment=PORT=8083
Environment=DB_HOST=localhost
Environment=DB_PORT=3306
Environment=DB_NAME=dev_workflow
Environment=DB_USERNAME=root
Environment=DB_PASSWORD=
Environment=RABBITMQ_URL=amqp://board:examplepassword@13.203.37.93:5672/
Environment=RABBITMQ_QUEUE_NAME=workflow_queue
Environment=WORKING_DIR=/tmp/claude-tasks

[Install]
WantedBy=multi-user.target
EOF

# 권한 설정
sudo chown -R ubuntu:ubuntu /opt/workflow

# systemd 리로드
sudo systemctl daemon-reload

echo "✅ Services configured successfully!"
echo "To start services manually:"
echo "  sudo systemctl start workflow-backend"
echo "  sudo systemctl start workflow-local-backend"
echo ""
echo "To enable services on boot:"
echo "  sudo systemctl enable workflow-backend"
echo "  sudo systemctl enable workflow-local-backend"