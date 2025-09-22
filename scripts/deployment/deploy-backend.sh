#!/bin/bash

# 백엔드 EC2 배포 스크립트
# 사용법: ./deploy-backend.sh [backend|local-backend] [environment]

set -e

COMPONENT=${1:-backend}
ENVIRONMENT=${2:-staging}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🚀 Starting $COMPONENT deployment to $ENVIRONMENT..."

# 환경별 설정
case $ENVIRONMENT in
  staging)
    EC2_HOST=${EC2_STAGING_HOST:-""}
    EC2_USER=${EC2_STAGING_USER:-"ubuntu"}
    ;;
  production)
    EC2_HOST=${EC2_PRODUCTION_HOST:-""}
    EC2_USER=${EC2_PRODUCTION_USER:-"ubuntu"}
    ;;
  *)
    echo "❌ Invalid environment: $ENVIRONMENT"
    echo "Usage: $0 [backend|local-backend] [staging|production]"
    exit 1
    ;;
esac

if [ -z "$EC2_HOST" ]; then
  echo "❌ EC2_HOST not configured for $ENVIRONMENT environment"
  exit 1
fi

# 컴포넌트별 설정
case $COMPONENT in
  backend)
    BUILD_DIR="$PROJECT_ROOT/backend"
    BINARY_NAME="server"
    SERVICE_NAME="workflow-backend"
    BUILD_CMD="go build -o $BINARY_NAME main.go"
    ;;
  local-backend)
    BUILD_DIR="$PROJECT_ROOT/local-backend"
    BINARY_NAME="local-server"
    SERVICE_NAME="workflow-local-backend"
    BUILD_CMD="go build -o $BINARY_NAME cmd/server/main.go"
    ;;
  *)
    echo "❌ Invalid component: $COMPONENT"
    echo "Usage: $0 [backend|local-backend] [staging|production]"
    exit 1
    ;;
esac

echo "📦 Building $COMPONENT..."
cd "$BUILD_DIR"

# 의존성 다운로드
go mod download

# 빌드
eval $BUILD_CMD

echo "📤 Deploying to EC2: $EC2_HOST"

# 바이너리 업로드
scp $BINARY_NAME $EC2_USER@$EC2_HOST:/tmp/

# EC2에서 배포 실행
ssh $EC2_USER@$EC2_HOST << EOF
  # 서비스 중지
  sudo systemctl stop $SERVICE_NAME || true
  
  # 백업 생성
  sudo cp /opt/workflow/$BINARY_NAME /opt/workflow/${BINARY_NAME}.backup.\$(date +%Y%m%d-%H%M%S) || true
  
  # 새 바이너리 배포
  sudo mkdir -p /opt/workflow
  sudo mv /tmp/$BINARY_NAME /opt/workflow/
  sudo chmod +x /opt/workflow/$BINARY_NAME
  sudo chown root:root /opt/workflow/$BINARY_NAME
  
  # 서비스 시작
  sudo systemctl start $SERVICE_NAME
  sudo systemctl enable $SERVICE_NAME
  
  # 상태 확인
  sleep 2
  sudo systemctl status $SERVICE_NAME --no-pager
EOF

echo "✅ $COMPONENT deployment completed!"
echo "🔍 Check service status: ssh $EC2_USER@$EC2_HOST 'sudo systemctl status $SERVICE_NAME'"