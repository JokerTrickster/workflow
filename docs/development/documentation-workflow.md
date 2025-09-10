# Documentation Workflow

## 📝 Claude + Obsidian MCP 문서화 워크플로우

### 🎯 목표
Claude AI와 Obsidian MCP 연동으로 프로젝트 문서를 자동화하고 체계적으로 관리

### 🔄 워크플로우

#### 1. **컴포넌트 문서화**
```bash
# Claude에게 요청
"src/presentation/components/TaskTab.tsx 컴포넌트의 기술문서를 작성해서 
docs/components/TaskTab.md에 저장해줘. 다음 내용을 포함해줘:
- 컴포넌트 개요
- Props 인터페이스
- 사용 예시
- 의존성 관계"
```

#### 2. **API 문서화**
```bash
# Claude에게 요청  
"GitHub API 연동 부분을 분석해서 API 문서를 작성해줘.
docs/api/github-integration.md에 저장하고 
엔드포인트, 요청/응답 형식, 에러 처리를 포함해줘"
```

#### 3. **아키텍처 문서화**
```bash
# Claude에게 요청
"전체 프로젝트 구조를 분석해서 mermaid 다이어그램으로 
아키텍처 문서를 만들어줘. 컴포넌트 관계도도 포함해서"
```

### 📋 문서화 체크리스트

#### ✅ 새 기능 개발 시
- [ ] 컴포넌트 문서 작성
- [ ] API 변경사항 문서 업데이트  
- [ ] 테스트 케이스 문서 추가
- [ ] README 업데이트

#### ✅ 리팩토링 시
- [ ] 아키텍처 문서 업데이트
- [ ] 컴포넌트 관계도 수정
- [ ] 의존성 문서 갱신

#### ✅ 버그 수정 시  
- [ ] 이슈 해결 과정 문서화
- [ ] 재발 방지 가이드 작성

### 🏷️ 문서 태그 시스템

```markdown
#component - 컴포넌트 관련
#api - API 관련  
#architecture - 아키텍처 관련
#workflow - 개발 워크플로우
#deployment - 배포 관련
#meeting - 미팅 노트
#decision - 의사결정 기록
#bug - 버그 관련
#feature - 기능 관련
```

### 📁 파일명 규칙

```
컴포넌트: ComponentName.md
API: api-endpoint-name.md  
아키텍처: system-architecture.md
워크플로우: workflow-name.md
미팅: YYYY-MM-DD-meeting-topic.md
```

### 🔗 링크 연결 규칙

- `[[Component Name]]` - 컴포넌트 간 연결
- `[[API Endpoint]]` - API 문서 연결  
- `[[System Overview]]` - 아키텍처 연결
- `![[diagram.png]]` - 이미지 임베드

### 💡 Claude 활용 팁

#### 문서 생성 요청 템플릿:
```
"[파일/컴포넌트명]에 대한 기술문서를 작성해줘. 
docs/[카테고리]/[파일명].md에 저장하고 다음을 포함해줘:
- 개요 및 목적
- 기술적 세부사항  
- 사용법 및 예시
- 관련 파일들과의 연결점
- 태그: #relevant-tags"
```

#### 다이어그램 생성 요청:
```
"[시스템/컴포넌트]의 구조를 mermaid 다이어그램으로 그려줘.
관계도와 데이터 플로우를 명확히 보여주고 
docs/architecture/diagrams/에 저장해줘"
```