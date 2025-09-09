---
task_id: 4
title: Build Korean Localization Infrastructure
epic: dashboard-ux-improvements
status: backlog
priority: medium
estimated_hours: 20
dependencies: none
assignee: unassigned
created: 2025-09-09T05:16:17Z
github_issue: https://github.com/JokerTrickster/workflow/issues/40
---

# Task 4: Build Korean Localization Infrastructure

## Description
한글 메시지 시스템을 구축하여 애플리케이션 전반에서 일관된 한국어 지원을 제공합니다. react-i18next 대신 경량 커스텀 훅을 사용하여 성능 오버헤드 없이 필요한 기능만 구현합니다. 기존 AuthContext 패턴을 따라 I18nContext를 구성합니다.

## Acceptance Criteria
- [ ] useI18n 커스텀 훅으로 한글/영어 메시지 관리
- [ ] 사용자 언어 설정 localStorage 저장
- [ ] 동적 파라미터 지원 (예: "{{name}}님 안녕하세요")
- [ ] TypeScript 안전한 키 자동완성 지원
- [ ] 기존 로그 메시지 한글화 적용
- [ ] 번들 크기 최소화 (50KB 이하)

## Technical Approach
- **경량 I18n 시스템**: react-i18next 대신 커스텀 구현
- **기존 패턴 활용**: AuthContext와 동일한 구조 적용
- **TypeScript 지원**: 메시지 키 타입 안전성 보장
- **Context API 사용**: I18nProvider로 전역 상태 관리

```typescript
// I18nContext 인터페이스
interface I18nContext {
  locale: 'ko' | 'en';
  t: (key: string, params?: Record<string, any>) => string;
  setLocale: (locale: 'ko' | 'en') => void;
}

// 사용법 예시
const { t } = useI18n();
const message = t('dashboard.repository.connected', { name: repoName });
```

## Dependencies
- 기존 Context 패턴 (AuthContext) 분석
- ActivityLogger의 메시지 구조 파악
- 한글화가 필요한 UI 텍스트 목록 작성

## Testing Requirements
- **단위 테스트**: useI18n 훅 기능 테스트
- **파라미터 테스트**: 동적 메시지 생성 검증
- **상태 테스트**: 언어 설정 저장/복원 테스트
- **통합 테스트**: 다양한 컴포넌트에서 한글 메시지 표시 검증
- **성능 테스트**: 번들 크기 및 런타임 성능 측정

## Definition of Done
- [ ] I18nProvider 및 useI18n 훅 구현 완료
- [ ] 한글/영어 메시지 사전 구축 완료
- [ ] TypeScript 타입 정의 완료
- [ ] 기존 UI 텍스트 한글화 적용
- [ ] 단위 테스트 및 통합 테스트 통과
- [ ] 성능 요구사항 (번들 크기) 충족