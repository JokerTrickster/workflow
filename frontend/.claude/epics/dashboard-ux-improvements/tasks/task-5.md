---
task_id: 5
title: Enhance Activity Logger with Korean Messages
epic: dashboard-ux-improvements
status: backlog
priority: medium
estimated_hours: 24
dependencies: 4
assignee: unassigned
created: 2025-09-09T05:16:17Z
github_issue: https://github.com/JokerTrickster/workflow/issues/41
---

# Task 5: Enhance Activity Logger with Korean Messages

## Description
기존 ActivityLogger 서비스를 확장하여 한글 메시지 지원을 추가하고, LogsTab에 필터링, 검색, 내보내기 기능을 구현합니다. 모든 사용자 액션과 GitHub API 호출을 한글로 기록하여 작업 내용을 쉽게 추적할 수 있도록 합니다.

## Acceptance Criteria
- [ ] 모든 사용자 액션 실시간 로그 기록 (한글)
- [ ] GitHub API 호출 상태를 한글로 로그 표시
- [ ] LogsTab에 날짜/시간별 필터링 기능 추가
- [ ] 로그 검색 기능 구현 (검색 응답시간 500ms 이하)
- [ ] 로그 데이터 내보내기 기능 (JSON, CSV)
- [ ] 로그 데이터 손실 0건 보장
- [ ] 실시간 로그 스트리밍 뷰어

## Technical Approach
- **기존 서비스 확장**: ActivityLogger에 한글 메시지 메서드 추가
- **I18n 통합**: Task 4의 useI18n 훅 활용
- **LogsTab 개선**: 기존 컴포넌트에 필터링/검색 UI 추가
- **실시간 업데이트**: useEffect + localStorage 변경 감지
- **성능 최적화**: 대량 로그 데이터 가상화

```typescript
// ActivityLogger 확장 예시
class ActivityLogger {
  logGitHubApiCall(endpoint: string, method: string, rateLimitRemaining: number) {
    const message = this.i18n.t('logs.github_api_call', {
      method,
      endpoint,
      rateLimitRemaining
    });
    this.addLog('api', message);
  }

  logUserAction(action: string, details?: Record<string, any>) {
    const message = this.i18n.t(`logs.user_actions.${action}`, details);
    this.addLog('user', message);
  }
}
```

## Dependencies
- Task 4 (Korean Localization Infrastructure) 완료 필요
- 기존 ActivityLogger 서비스 분석
- LogsTab 컴포넌트 구조 파악

## Testing Requirements
- **로그 기록 테스트**: 다양한 액션의 한글 로그 생성 검증
- **필터링 테스트**: 날짜/시간/타입별 로그 필터링 검증
- **검색 성능 테스트**: 대량 로그에서 검색 속도 측정
- **데이터 무결성 테스트**: 로그 데이터 손실 없음 검증
- **내보내기 테스트**: JSON, CSV 형식 내보내기 검증
- **실시간 업데이트 테스트**: 로그 실시간 스트리밍 검증

## Definition of Done
- [ ] ActivityLogger 한글 메시지 지원 구현 완료
- [ ] LogsTab 필터링/검색 기능 구현 완료
- [ ] 로그 내보내기 기능 구현 완료
- [ ] 실시간 로그 스트리밍 구현 완료
- [ ] 성능 요구사항 (검색 응답시간) 충족
- [ ] 단위 테스트 및 통합 테스트 통과