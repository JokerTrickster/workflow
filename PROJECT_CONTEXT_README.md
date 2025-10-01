# 프로젝트 컨텍스트 관리 시스템

## 개요

각 레포지토리에서 AI 작업 시 프로젝트별 컨텍스트와 규칙을 자동으로 로드하여 일관성 있는 작업을 수행합니다.

## 자동 생성 파일

모든 레포지토리에서 첫 작업 시 `.claude` 디렉토리와 다음 파일들이 자동 생성됩니다:

### 1. `.claude/PROJECT_CONTEXT.md`
프로젝트의 아키텍처, 코드베이스 구조, 중요한 규칙을 문서화합니다.

**위치**: `{repository}/.claude/PROJECT_CONTEXT.md`

**포함 내용**:
- 프로젝트 개요
- 아키텍처 설명
- 디렉토리 구조
- 주요 기술 스택
- 코딩 규칙
- 중요 파일 목록
- 의존성 정보
- 개발 워크플로우

### 2. `.claude/RULES.md`
프로젝트별 AI 작업 규칙과 제약사항을 정의합니다.

**위치**: `{repository}/.claude/RULES.md`

**포함 내용**:
- 코드 스타일
- 파일 구성 규칙
- 네이밍 규칙
- 테스트 요구사항
- 문서화 표준
- 금지 사항
- 필수 작업 사항

## 작동 방식

### 1. 자동 생성
```go
// 첫 작업 시 자동으로 실행됨
contextManager := NewProjectContextManager(workingDir)
contextManager.EnsureClaudeDirectory()
contextManager.CreateDefaultProjectContext(repositoryName)
contextManager.CreateDefaultProjectRules(repositoryName)
```

### 2. 프롬프트 통합
```go
// AI 작업 시 자동으로 컨텍스트 로드
projectContext := contextManager.GetContextForPrompt()
// PROJECT_CONTEXT.md + RULES.md 내용이 프롬프트에 추가됨
```

### 3. 작업 흐름
```
1. 태스크 생성
   ↓
2. 레포지토리 working directory 확인
   ↓
3. .claude/PROJECT_CONTEXT.md 존재 확인
   - 없으면 자동 생성 (템플릿)
   ↓
4. .claude/RULES.md 존재 확인
   - 없으면 자동 생성 (템플릿)
   ↓
5. 두 파일의 내용을 읽어서 프롬프트에 추가
   ↓
6. AI가 컨텍스트를 참고하여 작업 수행
```

## 사용 방법

### 초기 설정 (프로젝트마다 한 번만)

1. **첫 작업 실행**
   - 자동으로 `.claude` 디렉토리와 템플릿 파일 생성됨

2. **PROJECT_CONTEXT.md 작성**
   ```bash
   cd /path/to/your/repository
   vim .claude/PROJECT_CONTEXT.md
   ```

   프로젝트의 실제 정보로 템플릿을 채워넣습니다:
   - 아키텍처 다이어그램
   - 주요 컴포넌트 설명
   - 디렉토리별 역할
   - 핵심 기술 스택

3. **RULES.md 커스터마이징**
   ```bash
   vim .claude/RULES.md
   ```

   프로젝트별 규칙을 추가합니다:
   - 코드 스타일 가이드
   - 파일 생성 규칙
   - 테스트 작성 규칙
   - 금지/필수 사항

### 일상 사용

1. **자동 적용**
   - 모든 AI 작업에서 자동으로 컨텍스트 로드
   - 별도 설정 불필요

2. **컨텍스트 업데이트**
   ```bash
   # 프로젝트 구조 변경 시
   vim .claude/PROJECT_CONTEXT.md

   # 규칙 추가/변경 시
   vim .claude/RULES.md
   ```

3. **버전 관리**
   ```bash
   # .claude 디렉토리를 Git에 커밋
   git add .claude/
   git commit -m "docs: Update project context and rules"
   ```

## 예시

### PROJECT_CONTEXT.md 예시
```markdown
# gallery_ios - Project Context

## Project Overview
iOS 앱으로 S3 클라우드 스토리지에 사진/동영상을 업로드하고 관리하는 갤러리 애플리케이션

## Architecture
- MVVM 아키텍처
- Swift 5.0
- UIKit 기반

## Directory Structure
gallery_ios/
├── Models/          # 데이터 모델
├── Views/           # UI 컴포넌트
├── ViewModels/      # 비즈니스 로직
├── Services/        # API 통신
└── Utils/           # 유틸리티

## Key Technologies
- Swift 5.0
- UIKit
- AWS SDK for iOS
- Alamofire
```

### RULES.md 예시
```markdown
# gallery_ios - Project Rules

## Code Style
- Swift 표준 네이밍 규칙 사용
- camelCase for variables and functions
- PascalCase for classes and structs

## File Organization
- 모든 새 파일은 해당 역할의 디렉토리에 생성
- ViewModel은 반드시 ViewModels/ 디렉토리
- View는 Views/ 디렉토리

## Testing Requirements
- 모든 ViewModel은 Unit Test 필수
- 테스트 파일은 Tests/ 디렉토리

## Prohibited Actions
- UIViewController에 직접 비즈니스 로직 작성 금지
- 하드코딩된 API URL 사용 금지
```

## 레포지토리별 규칙 관리

### 전역 규칙 vs 프로젝트별 규칙

```
~/.claude/CLAUDE.md        # 전역 규칙 (모든 프로젝트)
    ↓
{repo}/.claude/RULES.md    # 프로젝트별 규칙 (해당 프로젝트만)
```

**우선순위**: 프로젝트별 규칙 > 전역 규칙

### 규칙 충돌 시
- 프로젝트별 RULES.md가 전역 설정을 오버라이드
- 명시적으로 정의된 규칙이 우선

## 구현 세부사항

### 파일 경로
```go
const (
    ClaudeDir          = ".claude"
    ProjectContextFile = "PROJECT_CONTEXT.md"
    ProjectRulesFile   = "RULES.md"
)
```

### 주요 함수
```go
// 프로젝트 컨텍스트 매니저 생성
contextManager := NewProjectContextManager(workingDir)

// .claude 디렉토리 생성
contextManager.EnsureClaudeDirectory()

// 파일 존재 확인
exists := contextManager.ProjectContextExists()

// 컨텍스트 읽기
context, err := contextManager.ReadProjectContext()

// 기본 파일 생성
contextManager.CreateDefaultProjectContext(repoName)

// 프롬프트용 컨텍스트 가져오기
promptContext := contextManager.GetContextForPrompt()
```

## 장점

1. **일관성**: 모든 프로젝트에서 동일한 위치에 문서 저장
2. **자동화**: 첫 작업 시 자동 생성, 수동 설정 불필요
3. **컨텍스트 유지**: AI가 항상 프로젝트 구조를 인식
4. **팀 협업**: Git으로 컨텍스트와 규칙 공유
5. **확장성**: 프로젝트별 커스터마이징 가능

## 체크리스트

새 프로젝트 시작 시:
- [ ] 첫 작업 실행 (자동 생성 트리거)
- [ ] `.claude/PROJECT_CONTEXT.md` 내용 작성
- [ ] `.claude/RULES.md` 프로젝트 규칙 추가
- [ ] `.claude/` 디렉토리를 Git에 커밋
- [ ] 팀원들과 문서 공유

기존 프로젝트에서:
- [ ] 구조 변경 시 PROJECT_CONTEXT.md 업데이트
- [ ] 새 규칙 추가 시 RULES.md 업데이트
- [ ] 정기적으로 문서 검토 및 개선

## 문제 해결

### Q: 파일이 자동 생성되지 않아요
A: `workingDir`이 올바르게 설정되었는지 확인하세요. 레포지토리 루트 디렉토리여야 합니다.

### Q: 컨텍스트가 로드되지 않아요
A: `.claude/PROJECT_CONTEXT.md` 파일이 UTF-8 인코딩인지 확인하세요.

### Q: 다른 레포지토리에도 적용되나요?
A: 네, 모든 레포지토리에서 자동으로 작동합니다. 각 레포지토리마다 별도의 `.claude/` 디렉토리가 생성됩니다.

### Q: 기존 코드베이스 분석은?
A: 첫 작업 시 AI에게 "현재 코드베이스를 분석해서 .claude/PROJECT_CONTEXT.md를 자세히 작성해줘"라고 요청하세요.

## 관련 파일

- `local-backend/utils/project_context.go` - 컨텍스트 관리 로직
- `local-backend/utils/claude_cli.go` - 프롬프트 통합
- `{repository}/.claude/PROJECT_CONTEXT.md` - 프로젝트 컨텍스트
- `{repository}/.claude/RULES.md` - 프로젝트 규칙
