# Workflow Backend API

클린아키텍처와 Echo 프레임워크를 사용한 Go 백엔드 API 서버입니다.

## 🏗️ 아키텍처

```
backend/
├── cmd/server/                 # 애플리케이션 엔트리포인트
├── internal/
│   ├── domain/                 # 비즈니스 로직 계층
│   │   ├── entities/          # 엔티티
│   │   ├── repositories/      # 리포지토리 인터페이스
│   │   └── services/          # 도메인 서비스
│   ├── usecase/               # 유즈케이스 계층
│   ├── delivery/              # 프레젠테이션 계층
│   │   └── http/
│   │       ├── handlers/      # HTTP 핸들러
│   │       ├── middleware/    # 미들웨어
│   │       └── routes/        # 라우트 설정
│   └── infrastructure/        # 인프라스트럭처 계층
│       ├── config/            # 설정
│       └── database/          # MySQL 데이터베이스 연결
├── bin/                       # 빌드된 바이너리
└── .env.example              # 환경변수 예제
```

## 🚀 시작하기

### 전제 조건
- Go 1.24.0 이상
- MySQL 8.0 이상
- Git

### MySQL 설정

1. **MySQL 서버 설치 및 실행**
```bash
# macOS (Homebrew)
brew install mysql
brew services start mysql

# Ubuntu
sudo apt update
sudo apt install mysql-server
sudo systemctl start mysql

# Windows
# MySQL Installer를 다운로드하여 설치
```

2. **데이터베이스 생성**
```sql
mysql -u root -p
CREATE DATABASE workflow CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'workflow_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON workflow.* TO 'workflow_user'@'localhost';
FLUSH PRIVILEGES;
```

### 설치 및 실행

1. **의존성 설치**
```bash
go mod tidy
```

2. **환경변수 설정**
```bash
cp .env.example .env
# .env 파일을 편집하여 MySQL 연결 정보를 설정
```

3. **서버 빌드**
```bash
go build -o bin/server cmd/server/main.go
```

4. **서버 실행**
```bash
./bin/server
```

또는 직접 실행:
```bash
go run cmd/server/main.go
```

## 📡 API 엔드포인트

### Health Check
- `GET /health` - 기본 헬스체크
- `GET /api/v1/health` - 상세 시스템 정보 포함 (메모리, 고루틴 등)
- `GET /api/v1/ping` - 간단한 핑/퐁 테스트

### Tasks
- `GET /api/v1/tasks` - 모든 태스크 조회
- `GET /api/v1/tasks/:id` - 특정 태스크 조회
- `POST /api/v1/tasks` - 새 태스크 생성
- `PUT /api/v1/tasks/:id` - 태스크 업데이트
- `DELETE /api/v1/tasks/:id` - 태스크 삭제

### Repositories
- `GET /api/v1/repositories` - 모든 저장소 조회
- `GET /api/v1/repositories/:id` - 특정 저장소 조회
- `POST /api/v1/repositories` - 새 저장소 연결
- `PUT /api/v1/repositories/:id` - 저장소 업데이트
- `DELETE /api/v1/repositories/:id` - 저장소 연결 해제

### GitHub Integration
- `POST /api/v1/github/webhook` - GitHub 웹훅 처리
- `GET /api/v1/github/repos` - GitHub 저장소 목록

### Workflows
- `GET /api/v1/workflows` - 워크플로우 목록
- `POST /api/v1/workflows` - 새 워크플로우 생성

## 🧪 API 테스트

```bash
# Health check
curl http://localhost:8080/health

# 상세 헬스체크 (시스템 정보 포함)
curl http://localhost:8080/api/v1/health | jq .

# Ping test
curl http://localhost:8080/api/v1/ping

# Tasks 조회
curl http://localhost:8080/api/v1/tasks | jq .

# Repositories 조회
curl http://localhost:8080/api/v1/repositories | jq .
```

## ⚙️ 환경변수

```bash
# 서버 설정
SERVER_PORT=8080
SERVER_HOST=localhost

# MySQL 데이터베이스 설정
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=workflow
DB_CHARSET=utf8mb4

# GitHub 설정 (향후 사용)
GITHUB_TOKEN=your_github_token
GITHUB_WEBHOOK_URL=https://your-domain.com/api/v1/github/webhook
```

## 🛠️ 기술 스택

- **프레임워크**: Echo v4.13.4
- **언어**: Go 1.24.0
- **데이터베이스**: MySQL 8.0+ (go-sql-driver/mysql)
- **아키텍처**: Clean Architecture
- **JSON 파싱**: 표준 json 패키지
- **로깅**: Echo의 내장 logger + 표준 log 패키지
- **미들웨어**: CORS, Logger, Recover

## 🗄️ 데이터베이스 설계

### 연결 풀 설정
- **MaxOpenConns**: 25개 연결
- **MaxIdleConns**: 25개 유휴 연결  
- **ConnMaxLifetime**: 5분

### 향후 테이블 구조 (예정)
```sql
-- 저장소 테이블
CREATE TABLE repositories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    description TEXT,
    private BOOLEAN DEFAULT FALSE,
    language VARCHAR(50),
    url VARCHAR(500),
    html_url VARCHAR(500),
    clone_url VARCHAR(500),
    stars INT DEFAULT 0,
    forks INT DEFAULT 0,
    is_connected BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 태스크 테이블
CREATE TABLE tasks (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status ENUM('pending', 'in_progress', 'completed', 'failed') DEFAULT 'pending',
    repository VARCHAR(255),
    epic VARCHAR(255),
    branch VARCHAR(255),
    tokens_used INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL
);
```

## 📝 향후 개발 계획

- [ ] MySQL 테이블 스키마 구현
- [ ] 데이터베이스 마이그레이션 시스템
- [ ] JWT 인증 시스템
- [ ] GitHub API 통합
- [ ] 웹소켓 지원 (실시간 알림)
- [ ] 로깅 시스템 개선 (구조화된 로깅)
- [ ] Docker 컨테이너화
- [ ] 유닛/통합 테스트 코드
- [ ] API 문서화 (Swagger/OpenAPI)
- [ ] 캐시 시스템 (Redis)
- [ ] 모니터링 및 메트릭

## 🔧 개발 명령어

```bash
# 의존성 설치
go mod tidy

# 개발 서버 실행
go run cmd/server/main.go

# 프로덕션 빌드
go build -o bin/server cmd/server/main.go

# 테스트 실행
go test ./...

# 린트 검사
golangci-lint run
```
