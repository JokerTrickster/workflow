---
task_id: 2
title: Implement Theme System with next-themes
epic: dashboard-ux-improvements
status: backlog
priority: high
estimated_hours: 28
dependencies: none
assignee: unassigned
created: 2025-09-09T05:16:17Z
github_issue: https://github.com/JokerTrickster/workflow/issues/38
---

# Task 2: Implement Theme System with next-themes

## Description
next-themes 라이브러리를 사용하여 시스템 전체 다크/라이트 모드를 구현합니다. 기존 AuthContext 패턴을 따라 ThemeProvider를 구성하고, 모든 기존 컴포넌트에 테마 지원을 추가합니다. 시스템 테마 자동 감지와 수동 토글 모두 지원합니다.

## Acceptance Criteria
- [ ] 시스템 테마 자동 감지 및 적용 (prefers-color-scheme)
- [ ] Header에 테마 전환 토글 버튼 추가
- [ ] 사용자별 테마 설정 localStorage 저장
- [ ] 새로고침 후에도 테마 설정 유지
- [ ] 모든 기존 컴포넌트에서 다크모드 정상 작동
- [ ] 테마 전환 속도 200ms 이하
- [ ] 테마 전환 시 부드러운 애니메이션 적용

## Technical Approach
- **next-themes 설치 및 설정**: _app.tsx에 ThemeProvider 통합
- **기존 패턴 활용**: AuthContext와 동일한 구조로 useTheme 훅 제공
- **Tailwind 다크모드**: tailwind.config.js에 dark mode 클래스 전략 적용
- **CSS Variables 확장**: 기존 스타일에 다크모드 색상 변수 추가
- **Header 컴포넌트 수정**: 테마 토글 버튼 추가 (기존 버튼 스타일 재사용)
- **점진적 적용**: 각 컴포넌트별로 dark: 클래스 적용

## Dependencies
- next-themes 라이브러리 설치
- 기존 컴포넌트 스타일 분석 완료
- Tailwind CSS 설정 이해

## Testing Requirements
- **시각적 테스트**: 라이트/다크 모드 스타일 검증
- **상태 테스트**: localStorage 저장/복원 테스트
- **성능 테스트**: 테마 전환 속도 측정
- **크로스 브라우저 테스트**: 다양한 브라우저에서 테마 동작 검증
- **단위 테스트**: useTheme 훅 및 ThemeToggle 컴포넌트 테스트

## Definition of Done
- [ ] next-themes 설치 및 ThemeProvider 설정 완료
- [ ] 모든 주요 컴포넌트에 다크모드 스타일 적용
- [ ] Header에 테마 토글 기능 통합
- [ ] 테마 전환 성능 요구사항 충족
- [ ] 단위 테스트 및 시각적 테스트 통과
- [ ] 코드 리뷰 완료 및 승인