# Project Standards Management Scripts

## Overview
이 디렉토리는 프로젝트 표준화 시스템을 위한 자동화 스크립트들을 포함합니다.

## Scripts

### 🚀 standards-cli.sh
**통합 CLI 인터페이스** - 모든 표준화 도구를 하나의 명령어로 사용

```bash
# 사용법
./standards-cli.sh <command> [options]

# 명령어
./standards-cli.sh init /path/to/project [name] [org]     # 새 프로젝트 초기화
./standards-cli.sh sync /path/to/project [--force]       # 기존 프로젝트 동기화
./standards-cli.sh validate /path/to/project [--fix]     # 프로젝트 검증
./standards-cli.sh status /path/to/project               # 프로젝트 상태 확인
./standards-cli.sh list                                  # 관리 중인 프로젝트 목록
./standards-cli.sh update                                # 템플릿 업데이트

# 대화형 모드
./standards-cli.sh  # 메뉴에서 선택
```

### 🛠️ init-project.sh
**새 프로젝트 초기화** - 표준 템플릿으로 새 프로젝트 설정

```bash
# 기본 사용법
./init-project.sh /path/to/new-project

# 상세 설정
./init-project.sh /path/to/project "My Project" "my-org"

# 생성되는 것들:
# - CLAUDE.md (프로젝트 표준 룰)
# - .gitignore
# - README.md
# - 기본 디렉토리 구조 (docs/, tests/, scripts/)
# - Git 저장소 초기화
```

### 🔄 sync-standards.sh
**기존 프로젝트 동기화** - 최신 표준으로 업데이트

```bash
# 기본 동기화 (확인 후 적용)
./sync-standards.sh /path/to/project

# 강제 동기화 (확인 없이 적용)
./sync-standards.sh /path/to/project --force

# 드라이런 (실제 변경 없이 미리보기)
./sync-standards.sh /path/to/project --dry-run

# 특징:
# - 기존 프로젝트별 룰 보존
# - 변경사항 백업 (.standards-backup/)
# - 충돌 시 수동 병합 지원
```

### ✅ validate-standards.sh
**프로젝트 표준 검증** - 룰 준수 여부 확인

```bash
# 기본 검증
./validate-standards.sh /path/to/project

# 자동 수정
./validate-standards.sh /path/to/project --fix

# 상세 리포트 생성
./validate-standards.sh /path/to/project --report

# 검증 항목:
# - CLAUDE.md 구조 및 내용
# - Git workflow 준수
# - 코드 품질 (TODO, 디버깅 문장 등)
# - 파일 구조 (테스트 위치 등)
# - 보안 (하드코딩된 비밀 등)
# - 패키지 스크립트
```

## 사용 시나리오

### 1. 새 프로젝트 시작
```bash
# 1. 프로젝트 초기화
./standards-cli.sh init /path/to/new-project my-app my-org

# 2. 상태 확인
./standards-cli.sh status /path/to/new-project

# 3. 개발 시작 (feature 브랜치)
cd /path/to/new-project
git checkout -b feature/initial-setup
```

### 2. 기존 프로젝트 표준화
```bash
# 1. 드라이런으로 변경사항 확인
./sync-standards.sh /path/to/existing-project --dry-run

# 2. 표준 적용
./sync-standards.sh /path/to/existing-project

# 3. 검증 및 수정
./validate-standards.sh /path/to/existing-project --fix

# 4. 상태 확인
./standards-cli.sh status /path/to/existing-project
```

### 3. 정기적인 표준 유지
```bash
# 1. 모든 프로젝트 목록 확인
./standards-cli.sh list

# 2. 각 프로젝트 검증
./standards-cli.sh validate /path/to/project1
./standards-cli.sh validate /path/to/project2

# 3. 필요시 동기화
./standards-cli.sh sync /path/to/project1 --force
```

## 자동화 설정

### Bash 별명 (alias) 설정
```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
alias pstd="/Users/luxrobo/project/workflow/.claude/scripts/standards/standards-cli.sh"

# 사용 예
pstd init /path/to/project
pstd status /path/to/project
pstd validate /path/to/project --fix
```

### 주기적 검증 (cron job)
```bash
# crontab -e로 추가
# 매주 월요일 오전 9시에 모든 프로젝트 검증
0 9 * * 1 /Users/luxrobo/project/workflow/.claude/scripts/standards/validate-all-projects.sh
```

### Git Hook 설정
프로젝트의 `.git/hooks/pre-commit`에 추가:
```bash
#!/bin/bash
# 커밋 전 표준 검증
/Users/luxrobo/project/workflow/.claude/scripts/standards/validate-standards.sh . || exit 1
```

## 출력 및 로깅

### 색상 코드
- 🔵 **[INFO]**: 일반 정보
- 🟢 **[SUCCESS]**: 성공적인 작업
- 🟡 **[WARNING]**: 주의사항 (진행 가능)
- 🔴 **[ERROR]**: 오류 (수정 필요)
- 🟢 **[FIXED]**: 자동 수정 완료

### 생성되는 파일들
- `.standards-backup/`: 변경 전 백업 파일들
- `standards-validation-report.md`: 상세 검증 리포트

## 문제 해결

### 권한 오류
```bash
chmod +x /Users/luxrobo/project/workflow/.claude/scripts/standards/*.sh
```

### 템플릿 경로 오류
스크립트들은 상대 경로를 사용합니다. workflow 프로젝트에서 실행해야 합니다.

### Git 오류
```bash
# Git 사용자 정보 설정
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### 변수 치환 오류
URL에 특수 문자가 포함된 경우, 자동으로 이스케이프 처리됩니다.

## 확장 및 커스터마이징

### 새로운 검증 규칙 추가
`validate-standards.sh`의 `validate_custom()` 함수에 추가:

```bash
validate_custom() {
    local project_path="$1"
    # 커스텀 검증 로직
}
```

### 새로운 템플릿 추가
`.claude/templates/` 디렉토리에 새 템플릿 추가 후 스크립트 수정

### 프로젝트별 커스터마이징
각 프로젝트의 `.claude/` 디렉토리에 커스텀 설정 파일 추가 가능

---

*이 스크립트들은 workflow 프로젝트 표준화 시스템의 일부입니다.*