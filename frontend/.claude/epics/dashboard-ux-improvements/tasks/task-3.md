---
task_id: 3
title: Implement Korean Comment System for GitHub Issues
epic: dashboard-ux-improvements
status: backlog
priority: high
estimated_hours: 36
dependencies: none
assignee: unassigned
created: 2025-09-09T05:16:17Z
github_issue: https://github.com/JokerTrickster/workflow/issues/39
---

# Task 3: Implement Korean Comment System for GitHub Issues

## Description
GitHub 이슈 작업 시 한글 코멘트를 자동으로 생성하는 시스템을 구현합니다. /pm:issue-start 명령어와 연동하여 작업 시작, 진행 상황, 완료 시점에 적절한 한글 코멘트를 GitHub 이슈에 자동으로 작성합니다. 기존 GitHubApiService를 확장하여 구현합니다.

## Acceptance Criteria
- [ ] /pm:issue-start 실행 시 "작업을 시작합니다" 한글 코멘트 생성
- [ ] 작업 진행 상황 업데이트 시 한글 상태 코멘트 작성
- [ ] 작업 완료 시 "작업이 완료되었습니다" 한글 코멘트 생성
- [ ] GitHub API 호출 실패 시 적절한 에러 핸들링
- [ ] 코멘트 템플릿 커스터마이징 가능
- [ ] ActivityLogger에 한글 코멘트 작성 로그 기록

## Technical Approach
- **기존 서비스 확장**: GitHubApiService에 createIssueComment 메서드 추가
- **한글 템플릿 시스템**: 작업 상태별 한글 메시지 템플릿 관리
- **PM 명령어 연동**: /pm:issue-start 처리 로직에 코멘트 기능 통합
- **에러 복구**: GitHub API rate limit 및 권한 에러 처리
- **ActivityLogger 확장**: 기존 서비스에 한글 메시지 지원 추가

```typescript
// 확장할 GitHubApiService 메서드
static async createIssueComment(
  repoId: string, 
  issueNumber: number, 
  comment: string
): Promise<void>

// 한글 템플릿 예시
const KOREAN_TEMPLATES = {
  start: "🚀 작업을 시작합니다.\n\n이 이슈 해결을 위한 작업을 시작하겠습니다.",
  progress: "⏳ 작업이 진행 중입니다.\n\n현재 상황: {{status}}",
  complete: "✅ 작업이 완료되었습니다.\n\n구현 내용을 검토해주세요."
}
```

## Dependencies
- 기존 GitHubApiService 코드 분석
- GitHub Issues API 권한 확인
- /pm:issue-start 명령어 처리 로직 파악

## Testing Requirements
- **API 통합 테스트**: GitHub Issues API 코멘트 작성 검증
- **에러 핸들링 테스트**: rate limit, 권한 에러 시나리오 검증  
- **템플릿 테스트**: 다양한 상황별 한글 메시지 생성 테스트
- **단위 테스트**: createIssueComment 메서드 및 템플릿 유틸리티
- **실제 시나리오 테스트**: /pm:issue-start 전체 플로우 검증

## Definition of Done
- [ ] GitHubApiService에 코멘트 기능 구현 완료
- [ ] 한글 템플릿 시스템 구축 완료
- [ ] /pm:issue-start와 통합 테스트 완료
- [ ] GitHub API 에러 처리 구현 완료
- [ ] 단위 테스트 및 통합 테스트 통과
- [ ] 실제 GitHub 이슈에서 동작 검증 완료