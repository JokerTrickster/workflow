# Base Rules - 모든 프로젝트 공통 기본 룰

## Workflow Rules 🟡

### Branch Strategy
- **Feature Branch Required**: 항상 default 브랜치에서 feature 브랜치 생성 후 작업
- **Rationale**: 메인 브랜치 안정성 보장, 코드 리뷰 강제화
- **Implementation**: `git checkout -b feature/task-name`
- **Validation**: 직접 main/master 브랜치에 커밋하지 않았는지 확인
- **Examples**:
  ✅ `git checkout -b feature/auth-improvements`
  ❌ `git checkout main && git commit -m "fix"`

### Pull Request Process
- **PR Required**: 모든 변경사항은 PR을 통해서만 메인 브랜치에 병합
- **Rationale**: 코드 품질 검토, 변경사항 추적, 팀 협업 강화
- **Implementation**: GitHub/GitLab PR 생성 후 리뷰 받기
- **Validation**: 메인 브랜치에 직접 푸시 없이 PR을 통해서만 병합

### Commit Messages
- **Descriptive Messages**: 구체적이고 의미 있는 커밋 메시지 작성
- **Rationale**: 변경사항 이해도 향상, 히스토리 추적 용이성
- **Implementation**: `feat: add user authentication`, `fix: resolve login timeout issue`
- **Validation**: "fix", "update", "changes" 같은 모호한 메시지 금지
- **Examples**:
  ✅ `feat: implement JWT token validation middleware`
  ❌ `fix stuff`, `update code`

## Code Quality Rules 🔴

### No Partial Implementation
- **Complete Features**: 시작한 기능은 반드시 동작 가능한 상태까지 완성
- **Rationale**: 불완전한 코드로 인한 시스템 불안정 방지
- **Implementation**: 기능 구현 시작 전 완료 기준 정의, 단계별 완성
- **Validation**: TODO 주석, 미완성 함수, mock 데이터 없이 완성
- **Examples**:
  ✅ `function calculateTotal() { return price * tax; }`
  ❌ `function calculateTotal() { throw new Error("Not implemented"); }`

### No Code Duplication
- **Reuse Existing Code**: 기존 코드베이스 확인 후 중복 코드 방지
- **Rationale**: 유지보수성 향상, 버그 수정 시 일관성 보장
- **Implementation**: 새 함수 작성 전 기존 유사 함수 검색
- **Validation**: 동일하거나 유사한 로직을 가진 함수가 여러 개 존재하지 않음
- **Examples**:
  ✅ Import and use existing `validateEmail()` function
  ❌ Create another `checkEmailFormat()` function

### No Dead Code
- **Active Code Only**: 사용하지 않는 코드는 완전히 제거
- **Rationale**: 코드베이스 크기 최적화, 혼란 방지, 보안 위험 제거
- **Implementation**: 정기적으로 unused imports, functions, variables 제거
- **Validation**: 사용되지 않는 코드가 존재하지 않음
- **Examples**:
  ✅ Remove unused functions and imports
  ❌ Comment out unused code "for future use"

### Consistent Naming
- **Follow Project Patterns**: 기존 프로젝트의 네이밍 컨벤션 따르기
- **Rationale**: 코드 가독성 향상, 팀 협업 효율성 증대
- **Implementation**: 기존 파일들의 네이밍 패턴 분석 후 동일 패턴 적용
- **Validation**: camelCase, snake_case, kebab-case가 혼재하지 않음
- **Examples**:
  ✅ `getUserData()` (if project uses camelCase)
  ❌ `get_userData()` (mixing conventions)

## Safety Rules 🔴

### Framework Respect
- **Check Dependencies**: 라이브러리 사용 전 기존 의존성 확인
- **Rationale**: 의존성 충돌 방지, 프로젝트 일관성 유지
- **Implementation**: package.json, requirements.txt 등 확인 후 사용
- **Validation**: 새로운 라이브러리 추가 시 기존 의존성과 충돌 없음

### Pattern Adherence
- **Follow Existing Conventions**: 프로젝트의 기존 패턴과 컨벤션 준수
- **Rationale**: 코드 일관성, 학습 곡선 최소화
- **Implementation**: 기존 코드의 구조, 스타일, 아키텍처 패턴 분석 후 적용
- **Validation**: 새 코드가 기존 패턴과 일관성을 가짐

### Resource Management
- **Clean Up Resources**: 연결, 타이머, 이벤트 리스너 등 리소스 정리
- **Rationale**: 메모리 누수 방지, 시스템 안정성 확보
- **Implementation**: try-finally, cleanup 함수, proper disposal
- **Validation**: 사용한 리소스가 적절히 해제되었는지 확인
- **Examples**:
  ✅ `connection.close()` in finally block
  ❌ Leave database connections open

## Testing Rules 🟡

### Test Everything
- **Comprehensive Testing**: 모든 함수와 기능에 대한 테스트 작성
- **Rationale**: 버그 조기 발견, 리팩토링 시 안정성 보장
- **Implementation**: 새 함수 작성 시 함께 테스트 코드 작성
- **Validation**: 테스트 없는 함수가 존재하지 않음

### Real Tests
- **Accurate Testing**: mock 사용 최소화, 실제 사용 시나리오 반영
- **Rationale**: 실제 환경에서 발생할 수 있는 문제 미리 발견
- **Implementation**: 실제 데이터, 실제 API 호출로 테스트
- **Validation**: 테스트가 실제 사용 패턴을 반영함

### Test Organization
- **Structured Testing**: tests/, __tests__, test/ 디렉토리에 테스트 파일 배치
- **Rationale**: 테스트 코드 관리 용이성, 프로젝트 구조 일관성
- **Implementation**: 테스트 파일을 소스 코드와 분리하여 배치
- **Validation**: 테스트 파일이 적절한 디렉토리에 위치

### Verbose Testing
- **Detailed Test Output**: 디버깅에 도움이 되는 상세한 테스트 작성
- **Rationale**: 테스트 실패 시 빠른 문제 파악
- **Implementation**: 명확한 테스트 이름, 상세한 assertion 메시지
- **Validation**: 테스트 실패 시 원인을 쉽게 파악할 수 있음

## Professional Standards 🟡

### Honest Assessment
- **Realistic Evaluation**: 마케팅 언어 사용하지 않고 실제 trade-off 제시
- **Rationale**: 정확한 의사결정을 위한 객관적 정보 제공
- **Implementation**: "blazingly fast", "100% secure" 등 과장 표현 금지
- **Validation**: 기술적 주장에 대한 근거나 측정 가능한 지표 제시

### Clean Workspace
- **Organized Environment**: 임시 파일, 불필요한 파일 정리
- **Rationale**: 전문적인 작업 환경 유지, 실수로 커밋 방지
- **Implementation**: 작업 완료 후 temp 파일, debug 스크립트 등 정리
- **Validation**: .gitignore에 명시되지 않은 임시 파일들이 남아있지 않음

### Quality Gates
- **Pre-completion Validation**: 작업 완료 전 lint, typecheck 실행
- **Rationale**: 코드 품질 표준 유지, 런타임 에러 사전 방지
- **Implementation**: `npm run lint`, `npm run typecheck` 등 실행 후 완료
- **Validation**: 모든 품질 검사 통과 후 작업 완료 표시

## Task Management 🟡

### Planning Required
- **TodoWrite for Complex Tasks**: 3단계 이상의 작업은 TodoWrite 도구 사용
- **Rationale**: 복잡한 작업의 체계적 관리, 진행상황 추적
- **Implementation**: 작업 시작 전 단계별 할일 목록 작성
- **Validation**: 복잡한 작업에 대한 계획이 문서화되어 있음

### Evidence-Based Development
- **Verifiable Claims**: 모든 기술적 주장은 테스트나 문서를 통해 검증 가능
- **Rationale**: 객관적이고 신뢰할 수 있는 개발 프로세스
- **Implementation**: 성능 개선, 버그 수정 등의 주장에 대한 증거 제시
- **Validation**: 주장에 대한 검증 가능한 근거 존재