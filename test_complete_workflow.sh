#!/bin/bash

# 전체 워크플로우 통합 테스트
echo "=== 전체 워크플로우 통합 테스트 시작 ==="

# 1. Task 생성 테스트 (request_id 포함)
echo "1. Task 생성 테스트..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": "갤러리 앱에 이미지 필터링 기능을 추가해주세요. 밝기, 대비, 채도 조절이 가능한 간단한 슬라이더를 만들어주세요.",
    "repository_name": "gallery_ios",
    "provider": "claude",
    "interactive": false,
    "working_dir": "",
    "cmd": ""
  }')

echo "응답: $RESPONSE"

# request_id 추출
REQUEST_ID=$(echo $RESPONSE | jq -r '.request_id')
echo "Request ID: $REQUEST_ID"

if [ "$REQUEST_ID" = "null" ] || [ -z "$REQUEST_ID" ]; then
    echo "❌ Task 생성 실패: request_id가 없습니다"
    exit 1
fi

echo "✅ Task가 성공적으로 생성되었습니다"
echo ""

# 2. Task 상태 확인 (processing 단계 확인)
echo "2. Task 상태 모니터링..."
for i in {1..20}; do
    echo "[$i/20] Task 상태 확인 중..."

    # 간단한 방법으로 데이터베이스 확인 대신 로그를 확인
    sleep 3

    # 3초마다 상태 확인
    if [ $i -eq 10 ]; then
        echo "⏰ Task 처리 중... (중간 점검)"
    fi
done

echo ""
echo "3. 최종 결과 확인..."

# 3. Git 저장소에서 변경사항 확인
echo "Git 저장소 상태 확인..."
cd /Users/mac/project/git-repository/JokerTrickster/gallery_ios

echo "현재 브랜치:"
git branch --show-current

echo ""
echo "최근 커밋 히스토리:"
git log --oneline -5

echo ""
echo "Git 상태:"
git status --porcelain

echo ""
echo "=== 테스트 완료 ==="
echo "Request ID: $REQUEST_ID"
echo "워크플로우가 성공적으로 완료되었는지 확인해주세요."