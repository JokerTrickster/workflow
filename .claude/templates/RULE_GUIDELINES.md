# Rule Addition Guidelines

## Overview
이 문서는 프로젝트 표준화 시스템에 새로운 룰을 추가하는 방법을 설명합니다.

## Rule Structure

### 1. Priority System 이해
```
🔴 CRITICAL   : 절대 타협불가 (보안, 데이터 안전, 운영 중단)
🟡 IMPORTANT  : 강한 선호도 (품질, 유지보수성, 전문성)
🟢 RECOMMENDED: 가능할 때 적용 (최적화, 스타일, 모범 사례)
```

### 2. Rule Template
```markdown
### {Category} Rules {Priority_Emoji}
- **{Rule Name}**: {명확한 설명}
- **Rationale**: {이 룰이 중요한 이유}
- **Implementation**: {적용 방법}
- **Validation**: {준수 확인 방법}
- **Examples**: {좋은 예시와 나쁜 예시}
```

## Category Guidelines

### Workflow Rules 🟡
**When to add**: 개발 프로세스, Git 워크플로우, 브랜치 전략과 관련된 룰
```markdown
### Workflow Rules 🟡
- **Feature Branch Naming**: feature/{type}-{description} 형식 사용
- **Rationale**: PR 추적과 자동화를 위한 일관성
- **Implementation**: `git checkout -b feature/auth-improvements`
- **Validation**: 브랜치명이 패턴과 일치하는지 확인
- **Examples**:
  ✅ `feature/auth-improvements`
  ❌ `fix-login-bug`
```

### Code Quality Rules 🔴/🟡
**When to add**: 코드 작성, 구조, 품질과 관련된 룰
```markdown
### Code Quality Rules 🔴
- **Error Handling**: 모든 async 함수는 try-catch 사용
- **Rationale**: 런타임 에러로 인한 서비스 중단 방지
- **Implementation**: Promise 기반 함수를 try-catch로 감싸기
- **Validation**: async 함수에 에러 핸들링이 있는지 확인
```

### Safety Rules 🔴
**When to add**: 보안, 데이터 안전, 시스템 안정성과 관련된 룰
```markdown
### Safety Rules 🔴
- **Environment Variables**: 민감 정보는 반드시 환경변수 사용
- **Rationale**: 코드 저장소에 민감 정보 노출 방지
- **Implementation**: .env 파일 사용, process.env.{VAR_NAME}
- **Validation**: 하드코딩된 API 키나 비밀번호가 없는지 확인
```

### Testing Rules 🟡
**When to add**: 테스트 작성, 실행, 구조와 관련된 룰
```markdown
### Testing Rules 🟡
- **Test Naming**: describe-it 패턴으로 행동 기반 테스트명 작성
- **Rationale**: 테스트 목적과 기대결과를 명확히 전달
- **Implementation**: describe("when user logs in", () => it("should return JWT token"))
- **Validation**: 테스트명이 행동과 결과를 설명하는지 확인
```

### Custom Categories
프로젝트 특화 룰은 도메인별로 분류:
- **API Rules**: REST API 설계, 응답 형식
- **Database Rules**: 쿼리 최적화, 트랜잭션 관리
- **UI Rules**: 컴포넌트 구조, 접근성
- **Performance Rules**: 최적화 기준, 메트릭

## Rule Writing Best Practices

### 1. 구체적이고 실행 가능하게
```markdown
❌ "코드를 깔끔하게 작성하세요"
✅ "함수는 단일 책임을 가져야 하며, 20줄을 초과하지 않아야 합니다"
```

### 2. 검증 가능하게
```markdown
❌ "성능을 고려하세요"
✅ "API 응답 시간은 200ms 이하여야 하며, lighthouse 스코어 90 이상 유지"
```

### 3. 예시 포함
```markdown
**Examples**:
✅ Good: `const getUserData = async (userId) => { try { ... } catch (error) { ... } }`
❌ Bad: `const getUserData = async (userId) => { return api.getUser(userId); }`
```

## Priority Decision Framework

### 🔴 CRITICAL 판단 기준
- 보안 취약점 발생 가능
- 데이터 손실이나 corruption 위험
- 서비스 중단으로 이어질 수 있음
- 법적/규정 준수 필요

### 🟡 IMPORTANT 판단 기준
- 코드 품질과 유지보수성에 큰 영향
- 팀 협업 효율성 향상
- 장기적 기술 부채 방지
- 전문적 개발 표준

### 🟢 RECOMMENDED 판단 기준
- 코드 가독성 향상
- 개발자 경험 개선
- 성능 최적화
- 업계 모범 사례

## Rule Validation

### 자동 검증 가능한 룰
```markdown
### Code Quality Rules 🔴
- **Linting**: ESLint/Prettier 규칙 준수
- **Implementation**: .eslintrc.js 설정 파일
- **Validation**: `npm run lint` 실행 시 에러 없음
```

### 수동 검증 필요한 룰
```markdown
### Workflow Rules 🟡
- **Code Review**: 모든 PR은 최소 1명의 리뷰어 승인 필요
- **Implementation**: GitHub branch protection rules
- **Validation**: PR merge 전 리뷰어 승인 확인
```

## 룰 추가 프로세스

1. **필요성 확인**: 왜 이 룰이 필요한가?
2. **우선순위 결정**: 🔴/🟡/🟢 중 선택
3. **카테고리 선택**: 적절한 카테고리 확인
4. **룰 작성**: 템플릿 따라 작성
5. **검증 방법 정의**: 어떻게 확인할 것인가?
6. **예시 추가**: 좋은/나쁜 예시 포함
7. **테스트**: 실제 프로젝트에 적용해보기
8. **문서화**: CLAUDE.md에 추가

## 실제 적용 예시

### 프로젝트별 커스텀 룰
```markdown
## Project-Specific Rules

### API Rules 🟡
- **Response Structure**: 모든 API는 {status, data, message} 구조 사용
- **Rationale**: 프론트엔드에서 일관된 에러 처리 가능
- **Implementation**: response.json({status: 'success', data: result, message: ''})
- **Validation**: API 테스트에서 응답 구조 확인
- **Examples**:
  ✅ `{status: 'success', data: {user}, message: 'User created'}`
  ❌ `{user: {}, success: true}`

### Database Rules 🔴
- **Connection Management**: 모든 DB 쿼리 후 연결 해제 필수
- **Rationale**: 커넥션 풀 고갈로 인한 서비스 중단 방지
- **Implementation**: try-finally 블록에서 connection.close() 호출
- **Validation**: 커넥션 리크 모니터링 도구로 확인
```

이렇게 구조화된 가이드라인으로 일관성 있고 확장 가능한 룰 시스템을 만들 수 있습니다.