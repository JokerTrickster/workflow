#!/bin/bash

# 프론트엔드 S3 배포 스크립트
# 사용법: ./deploy-frontend.sh [environment]

set -e

ENVIRONMENT=${1:-staging}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🚀 Starting frontend deployment to $ENVIRONMENT..."

# 환경별 설정
case $ENVIRONMENT in
  staging)
    S3_BUCKET=${S3_STAGING_BUCKET:-"workflow-frontend-staging"}
    CLOUDFRONT_ID=${CLOUDFRONT_STAGING_ID:-""}
    ;;
  production)
    S3_BUCKET=${S3_PRODUCTION_BUCKET:-"workflow-frontend-prod"}
    CLOUDFRONT_ID=${CLOUDFRONT_PRODUCTION_ID:-""}
    ;;
  *)
    echo "❌ Invalid environment: $ENVIRONMENT"
    echo "Usage: $0 [staging|production]"
    exit 1
    ;;
esac

echo "📦 Building frontend..."
cd "$PROJECT_ROOT/frontend"

# 의존성 설치
npm ci

# 환경별 빌드
if [ "$ENVIRONMENT" = "production" ]; then
  npm run build:prod
else
  npm run build
fi

echo "☁️ Uploading to S3 bucket: $S3_BUCKET"

# S3 업로드
aws s3 sync dist/ s3://$S3_BUCKET --delete --acl public-read

# CloudFront 캐시 무효화
if [ ! -z "$CLOUDFRONT_ID" ]; then
  echo "🔄 Invalidating CloudFront cache..."
  aws cloudfront create-invalidation \
    --distribution-id $CLOUDFRONT_ID \
    --paths "/*"
fi

echo "✅ Frontend deployment completed!"
echo "🌐 Website URL: https://$S3_BUCKET.s3-website-${AWS_REGION:-us-east-1}.amazonaws.com"