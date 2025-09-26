# Project Standards Templates

## Overview
이 디렉토리는 모든 프로젝트에서 사용할 표준화된 개발 룰과 설정 템플릿을 포함합니다.

## Template Structure
```
templates/
├── README.md              # 이 파일
├── CLAUDE.md              # 표준 프로젝트 룰 템플릿
├── RULE_GUIDELINES.md     # 룰 작성 가이드라인
├── rules/                 # 룰 시스템 아키텍처
│   ├── README.md          # 룰 시스템 설명서
│   ├── base-rules.md      # 모든 프로젝트 공통 기본 룰
│   └── rule-template.md   # 새 룰 작성 템플릿
├── configs/               # 설정 파일 템플릿
└── workflows/             # GitHub Actions 워크플로우 템플릿
```

## Usage

### 1. 새 프로젝트 초기화
```bash
# workflow 프로젝트로 이동
cd /Users/luxrobo/project/workflow

# 새 프로젝트에 표준 템플릿 적용 (향후 스크립트로 자동화)
cp .claude/templates/CLAUDE.md /path/to/new-project/CLAUDE.md
```

### 2. 기존 프로젝트 표준화
```bash
# 기존 CLAUDE.md와 템플릿 비교
diff existing-project/CLAUDE.md .claude/templates/CLAUDE.md

# 필요한 부분만 선별적으로 적용
```

### 3. 새 룰 추가
```bash
# 룰 가이드라인 참고
cat .claude/templates/RULE_GUIDELINES.md

# 룰 템플릿 복사하여 새 룰 작성
cp .claude/templates/rules/rule-template.md my-new-rule.md
```

## Template Customization

### CLAUDE.md 템플릿 변수
템플릿에서 다음 변수들을 실제 프로젝트 정보로 교체:
- `{REPOSITORY_NAME}`: GitHub 저장소 이름
- `{GITHUB_URL}`: GitHub 저장소 URL
- `{PROJECT_NAME}`: 프로젝트 이름
- `{LAST_UPDATED}`: 마지막 업데이트 날짜

### 프로젝트별 커스텀 룰 추가
CLAUDE.md 템플릿의 "Project-Specific Rules" 섹션에 다음과 같이 추가:

```markdown
## Project-Specific Rules

### API Rules 🟡
- **Response Format**: 모든 API는 {status, data, message} 구조 사용
- **Rationale**: 프론트엔드에서 일관된 에러 처리 가능
- **Implementation**: response.json({status: 'success', data: result})
- **Validation**: API 테스트에서 응답 구조 확인

### Database Rules 🔴
- **Connection Management**: 모든 DB 쿼리 후 연결 해제 필수
- **Rationale**: 커넥션 풀 고갈로 인한 서비스 중단 방지
- **Implementation**: try-finally 블록에서 connection.close() 호출
- **Validation**: 커넥션 리크 모니터링
```

## Quality Assurance

### 템플릿 검증 체크리스트
- [ ] 모든 룰에 우선순위 이모지 (🔴🟡🟢) 표시
- [ ] 룰별로 Rationale, Implementation, Validation 포함
- [ ] 좋은 예시/나쁜 예시 제공
- [ ] 프로젝트별 변수 ({}) 명확히 표시
- [ ] 확장 가능한 구조로 설계
- [ ] 기존 workflow 프로젝트 룰과 일관성 유지

### 적용 후 검증
- [ ] 새 프로젝트에 템플릿 적용 테스트
- [ ] 기존 프로젝트와의 호환성 확인
- [ ] 개발자 피드백 수집 및 반영
- [ ] 룰 준수 여부 모니터링

## Maintenance

### 정기 업데이트
- **월간**: 새로운 모범 사례 반영
- **분기별**: 프로젝트 피드백 기반 룰 개선
- **연간**: 전체 룰 시스템 리뷰 및 재구성

### 버전 관리
- 템플릿 변경 시 Git 커밋으로 변경 이력 관리
- 주요 변경 시 CHANGELOG.md 업데이트
- 기존 프로젝트 영향도 분석 후 마이그레이션 가이드 제공

## Next Steps

### 즉시 구현 가능
1. ✅ 표준 CLAUDE.md 템플릿 완성
2. ✅ 룰 작성 가이드라인 문서화
3. ✅ 기본 룰셋 정의

### 단계적 구현 예정
1. 🔄 프로젝트 초기화 스크립트 개발
2. ⏳ 기존 프로젝트 표준화 스크립트
3. ⏳ GitHub Actions 워크플로우 템플릿
4. ⏳ 자동 검증 도구 개발

### 장기 계획
1. 📋 웹 기반 룰 관리 인터페이스
2. 📋 프로젝트별 룰 준수 대시보드
3. 📋 AI 기반 룰 준수 자동 검증
4. 📋 팀별 커스텀 룰 템플릿