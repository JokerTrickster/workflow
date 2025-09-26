# Rules System Architecture

## Overview
확장 가능한 룰 시스템으로 프로젝트별 개발 표준을 체계적으로 관리합니다.

## Directory Structure
```
rules/
├── README.md              # 이 파일
├── base-rules.md         # 모든 프로젝트 공통 기본 룰
├── categories/           # 카테고리별 룰 모음
│   ├── workflow.md       # Git, 브랜치, PR 워크플로우
│   ├── code-quality.md   # 코딩 표준, 구조, 품질
│   ├── safety.md         # 보안, 데이터 안전, 시스템 안정성
│   ├── testing.md        # 테스트 전략, 구조, 실행
│   └── performance.md    # 성능 최적화, 모니터링
├── templates/            # 룰 작성 템플릿
│   └── rule-template.md  # 새 룰 작성 템플릿
├── validators/           # 룰 검증 스크립트
│   ├── lint-rules.js     # 자동 검증 가능한 룰
│   └── checklist.md      # 수동 검증 체크리스트
└── examples/             # 프로젝트별 실제 적용 예시
    ├── frontend-rules.md # 프론트엔드 프로젝트 예시
    ├── backend-rules.md  # 백엔드 프로젝트 예시
    └── fullstack-rules.md# 풀스택 프로젝트 예시
```

## Rule Priority System
- 🔴 **CRITICAL**: 절대 타협불가 (보안, 데이터 안전, 운영 중단)
- 🟡 **IMPORTANT**: 강한 선호도 (품질, 유지보수성, 전문성)
- 🟢 **RECOMMENDED**: 가능할 때 적용 (최적화, 스타일, 모범 사례)

## Rule Categories

### Core Categories
1. **Workflow**: Git 워크플로우, 브랜치 전략, PR 프로세스
2. **Code Quality**: 코딩 표준, 구조, 품질 관리
3. **Safety**: 보안, 데이터 보호, 시스템 안정성
4. **Testing**: 테스트 전략, 구조, 실행 방법
5. **Performance**: 성능 최적화, 모니터링

### Custom Categories
프로젝트별 도메인 특화 룰:
- **API Rules**: REST/GraphQL 설계, 응답 형식
- **Database Rules**: 쿼리 최적화, 트랜잭션 관리
- **UI Rules**: 컴포넌트 구조, 접근성, 디자인 시스템
- **DevOps Rules**: 배포, 모니터링, 인프라 관리

## Usage Pattern

### 1. 새 프로젝트 초기화
```bash
# 기본 룰셋 적용
./scripts/init-project.sh project-name

# 특정 카테고리 룰 추가
./scripts/add-rules.sh project-name api database
```

### 2. 기존 프로젝트에 룰 추가
```bash
# 새 룰 작성
cp templates/rule-template.md categories/my-new-rule.md

# 프로젝트에 적용
./scripts/sync-rules.sh project-name
```

### 3. 룰 검증
```bash
# 자동 검증
./scripts/validate-rules.sh project-name

# 수동 체크리스트
cat validators/checklist.md
```

## Rule Inheritance

### 계층 구조
```
Global Rules (base-rules.md)
    ↓
Category Rules (categories/*.md)
    ↓
Project-Specific Rules (CLAUDE.md)
```

### 우선순위
1. Project-Specific Rules (최우선)
2. Category Rules
3. Global Base Rules

## Extension Points

### 1. 새로운 카테고리 추가
```bash
# 새 카테고리 파일 생성
touch categories/my-category.md

# 템플릿 복사하여 편집
cp templates/rule-template.md categories/my-category.md
```

### 2. 검증 스크립트 추가
```bash
# 새 검증 스크립트
touch validators/my-validator.js

# package.json에 스크립트 추가
"validate:my-rules": "node validators/my-validator.js"
```

### 3. 프로젝트 예시 추가
```bash
# 새 프로젝트 타입 예시
touch examples/my-project-type-rules.md
```

## Best Practices

### Rule Writing
1. **구체적이고 실행 가능하게** 작성
2. **검증 가능한 기준** 제시
3. **좋은/나쁜 예시** 포함
4. **rationale 명확히** 설명

### Rule Organization
1. **단일 책임**: 한 룰은 한 가지만 다룸
2. **계층 구조**: 일반적 → 구체적 순서
3. **의존성 관리**: 룰 간 충돌 방지
4. **버전 관리**: 룰 변경 이력 추적

### Validation Strategy
1. **자동화 우선**: 가능한 한 자동 검증
2. **단계적 적용**: Critical → Important → Recommended
3. **점진적 도입**: 기존 프로젝트에 단계적 적용
4. **피드백 수집**: 룰 효과성 모니터링

## Migration Guide

### 기존 프로젝트 적용
1. **현재 상태 분석**: 기존 CLAUDE.md 검토
2. **우선순위 선별**: Critical 룰부터 적용
3. **단계적 마이그레이션**: 한 번에 모든 룰 적용하지 말고 점진적으로
4. **검증 및 수정**: 적용 후 문제점 파악하여 개선

### 새 프로젝트 적용
1. **프로젝트 타입 확인**: Frontend/Backend/Fullstack
2. **기본 룰셋 선택**: 해당하는 예시 참고
3. **커스텀 룰 추가**: 프로젝트 특성에 맞게 확장
4. **검증 환경 구성**: 자동 검증 스크립트 설정