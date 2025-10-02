# Docker 기반 백엔드 배포 가이드

## 개요

Docker와 GitHub Actions를 사용한 간단한 백엔드 배포 방법입니다.

## 사전 준비

### 1. EC2에 Docker 설치

```bash
# Docker 설치
sudo apt-get update
sudo apt-get install -y docker.io docker-compose

# Docker 서비스 시작
sudo systemctl start docker
sudo systemctl enable docker

# 현재 사용자를 docker 그룹에 추가 (sudo 없이 사용)
sudo usermod -aG docker $USER
newgrp docker

# 설치 확인
docker --version
docker-compose --version
```

### 2. GitHub Secrets 설정

Repository → Settings → Secrets and variables → Actions에서 설정:

- `EC2_HOST`: `13.203.37.93`
- `EC2_USER`: `ubuntu`
- `EC2_SSH_KEY`: SSH 개인키 전체 내용
- `GITHUB_TOKEN`: GitHub Personal Access Token

## 배포 방법

### GitHub Actions로 자동 배포

1. GitHub → Actions → "Deploy Backend" 선택
2. "Run workflow" 클릭
3. 배포할 브랜치 선택 (기본: main)
4. "Run workflow" 클릭

### 수동 배포 (로컬에서)

```bash
# 1. 코드를 EC2에 복사
scp -r backend ubuntu@13.203.37.93:/tmp/

# 2. EC2에 접속
ssh ubuntu@13.203.37.93

# 3. 배포
cd /tmp/backend
export GITHUB_TOKEN=your_token_here
docker-compose down
docker-compose up -d --build

# 4. 상태 확인
docker-compose ps
docker-compose logs -f
```

## 배포 구조

```
backend/
├── Dockerfile              # 컨테이너 이미지 정의
├── docker-compose.yml      # 서비스 구성
├── main.go                 # 애플리케이션 엔트리포인트
└── ...
```

### Dockerfile
- **Build Stage**: Go 1.21로 바이너리 빌드
- **Runtime Stage**: Alpine Linux에서 경량 실행

### docker-compose.yml
- 포트: 7000
- 자동 재시작: always
- 환경변수: 모든 설정 포함
- 볼륨: 로그 및 작업 디렉토리 마운트

## 컨테이너 관리

### 로그 확인
```bash
docker-compose logs -f
docker-compose logs --tail=100
```

### 컨테이너 재시작
```bash
docker-compose restart
```

### 컨테이너 중지
```bash
docker-compose down
```

### 컨테이너 재빌드
```bash
docker-compose up -d --build
```

### 상태 확인
```bash
docker-compose ps
docker ps
```

## 문제 해결

### 컨테이너가 시작되지 않음
```bash
# 로그 확인
docker-compose logs

# 개별 컨테이너 로그
docker logs workflow-backend

# 컨테이너 상세 정보
docker inspect workflow-backend
```

### 포트 충돌
```bash
# 7000번 포트 사용 확인
sudo lsof -i :7000

# 기존 프로세스 종료
sudo kill -9 <PID>
```

### 이미지 정리
```bash
# 사용하지 않는 이미지 삭제
docker image prune -f

# 모든 중지된 컨테이너 삭제
docker container prune -f

# 전체 정리 (주의!)
docker system prune -a -f
```

## 환경변수 수정

`docker-compose.yml` 파일에서 환경변수 수정 후:

```bash
docker-compose down
docker-compose up -d
```

## 백업 및 롤백

### 현재 이미지 백업
```bash
docker tag workflow-backend workflow-backend:backup-$(date +%Y%m%d)
```

### 이전 버전으로 롤백
```bash
docker-compose down
docker tag workflow-backend:backup-20250102 workflow-backend:latest
docker-compose up -d
```

## 보안 권장사항

1. **환경변수 보호**: 민감한 정보는 `.env` 파일이나 Secrets 사용
2. **네트워크 격리**: 필요한 포트만 노출
3. **정기 업데이트**: 베이스 이미지 정기적 업데이트
4. **로그 관리**: 로그 파일 크기 제한 설정

## 모니터링

### 리소스 사용량 확인
```bash
docker stats workflow-backend
```

### 헬스체크
```bash
curl http://localhost:7000/health
```

## 장점

- ✅ **간단한 배포**: 코드 복사 → 컨테이너 재시작
- ✅ **환경 일관성**: 로컬/프로덕션 동일 환경
- ✅ **빠른 롤백**: 이전 이미지로 즉시 복구
- ✅ **격리된 실행**: 호스트 시스템과 독립적
- ✅ **자동 재시작**: 장애 시 자동 복구
