---
name: dashboard-ux-improvements
status: backlog
created: 2025-09-09T05:13:01Z
progress: 0%
prd: .claude/prds/dashboard-ux-improvements.md
github: Issues #37-#42 created
---

# Epic: Dashboard UX Improvements

## Overview

Dashboard UX 개선 에픽은 기존 Next.js + TypeScript 스택을 활용하여 4가지 핵심 UX 문제를 해결합니다: 1) 깃허브 연동 후 대시보드 스크롤 이슈 수정, 2) 이슈 작업 시 한글 코멘트 자동화, 3) 시스템 전체 다크모드 지원, 4) 향상된 활동 로그 시스템. 기존 아키텍처와 컴포넌트를 최대한 재사용하면서 점진적으로 개선하는 접근 방식을 채택합니다.

## Architecture Decisions

### Design Patterns & Approaches
- **Context API 활용**: 테마 상태 관리를 위해 기존 AuthContext 패턴을 확장한 ThemeContext 생성
- **Hook 패턴 확장**: 기존 useRepositories, useNetworkStatus 패턴을 따라 useTheme, useI18n 훅 추가
- **Error Boundary 재사용**: 기존 ErrorBoundary 컴포넌트를 활용한 안정적인 스크롤 처리
- **CSS Variables 활용**: Tailwind CSS dark: 클래스와 CSS 커스텀 프로퍼티를 조합한 테마 시스템

### Technology Choices
- **스크롤 최적화**: React-Query의 useInfiniteQuery + intersection Observer API
- **테마 시스템**: next-themes 라이브러리 + Tailwind CSS dark mode
- **한글화**: 경량 i18n 유틸리티 (react-i18next 대신 custom hook 사용)
- **로그 시스템**: 기존 ActivityLogger 서비스 확장

### Integration Strategy
- GitHub API 호출 패턴 유지하면서 코멘트 기능만 추가
- 기존 컴포넌트 구조 (Header, Dashboard, WorkspacePanel) 최소 변경
- Supabase Auth 연동 상태 그대로 유지

## Technical Approach

### Frontend Components

#### 스크롤 문제 해결
- **Dashboard.tsx 수정**: 무한 스크롤 로직 개선 및 virtual scrolling 적용
- **RepositoryCard 최적화**: React.memo 적용으로 렌더링 성능 개선
- **IntersectionObserver 훅**: useIntersectionObserver 커스텀 훅으로 스크롤 감지

#### 테마 시스템 구현
- **ThemeProvider 컴포넌트**: system/light/dark 모드 지원
- **ThemeToggle 컴포넌트**: Header에 통합되는 테마 전환 버튼
- **CSS Variables 확장**: 기존 Tailwind 설정에 다크모드 색상 변수 추가

#### 한글 코멘트 시스템
- **CommentService**: GitHub API 코멘트 작성 전용 서비스
- **I18nUtils**: 한글 메시지 템플릿 관리 유틸리티
- **IssueWorkflow 훅**: /pm:issue-start 명령어 처리 및 코멘트 자동화

#### 로그 개선
- **ActivityLogger 확장**: 기존 서비스에 한글 메시지 지원 추가
- **LogsTab 개선**: 필터링, 검색, 내보내기 기능 추가
- **LogViewer 컴포넌트**: 실시간 로그 스트리밍 뷰어

### State Management Approach

```typescript
// 테마 상태 관리
interface ThemeContext {
  theme: 'system' | 'light' | 'dark';
  setTheme: (theme: string) => void;
  resolvedTheme: 'light' | 'dark';
}

// 한글화 상태 관리  
interface I18nContext {
  locale: 'ko' | 'en';
  t: (key: string, params?: Record<string, any>) => string;
  setLocale: (locale: string) => void;
}
```

### User Interaction Patterns
- **Progressive Enhancement**: 기본 기능 우선, 고급 기능은 점진적 로드
- **Optimistic Updates**: 테마 전환, 코멘트 작성 시 즉시 UI 반영
- **Graceful Degradation**: JavaScript 비활성 시에도 기본 스크롤 동작 유지

## Implementation Strategy

### Development Phases

**Phase 1 (Week 1): Critical Fixes**
- 대시보드 스크롤 문제 근본 원인 분석 및 수정
- 기본 다크모드 구현 (시스템 테마 감지)
- 성능 모니터링 도구 설치

**Phase 2 (Week 2): Core Features** 
- 한글 코멘트 시스템 구현
- 완전한 테마 시스템 (수동 토글 포함)
- 기존 컴포넌트에 테마 적용

**Phase 3 (Week 3): Enhancement**
- 활동 로그 시스템 개선
- 모바일 최적화 및 성능 튜닝
- QA 및 버그 픽스

### Risk Mitigation
- **스크롤 이슈**: 여러 브라우저에서 원인 분석 우선, 폴백 메커니즘 준비
- **API 제한**: GitHub API 호출 최적화 및 rate limit 모니터링
- **테마 충돌**: 단계적 롤아웃으로 스타일 충돌 최소화

### Testing Approach
- **Unit Tests**: 새로운 훅과 유틸리티 함수 테스트
- **Integration Tests**: 스크롤, 테마 전환, 코멘트 작성 시나리오
- **E2E Tests**: 전체 사용자 여정 테스트
- **Performance Tests**: 스크롤 성능 및 메모리 사용량 모니터링

## Task Breakdown Preview

High-level task categories (총 8개 작업으로 제한):

- [ ] **Dashboard Scroll Fix**: 무한 스크롤 및 성능 최적화 (3-4일)
- [ ] **Theme System Core**: next-themes 통합 및 기본 다크모드 (2-3일) 
- [ ] **Theme UI Integration**: 모든 컴포넌트에 테마 적용 (3-4일)
- [ ] **Korean Comment Service**: GitHub API 코멘트 자동화 시스템 (4-5일)
- [ ] **I18n Infrastructure**: 한글 메시지 시스템 구축 (2-3일)
- [ ] **Enhanced Activity Logs**: 로그 필터링, 검색, 내보내기 (3-4일)
- [ ] **Mobile & Performance**: 모바일 최적화 및 성능 튜닝 (2-3일)
- [ ] **Testing & QA**: 통합 테스트 및 품질 보증 (3-4일)

## Dependencies

### External Dependencies
- **next-themes**: 테마 시스템 관리
- **@headlessui/react**: 테마 토글 UI 컴포넌트
- **intersection-observer polyfill**: 구형 브라우저 호환성
- **GitHub API v3**: 이슈 코멘트 작성 기능

### Internal Dependencies  
- **기존 GitHubApiService**: 코멘트 메서드 추가 확장
- **ActivityLogger**: 한글 메시지 지원 확장
- **useRepositories 훅**: 스크롤 최적화를 위한 수정
- **ErrorBoundary 컴포넌트**: 안정적인 스크롤 처리

### Team Dependencies
- **Frontend Team**: 모든 구현 담당
- **QA Team**: 다양한 브라우저/기기에서 테스트
- **DevOps Team**: 배포 시 환경 변수 설정 지원

## Success Criteria (Technical)

### Performance Benchmarks
- **스크롤 성능**: 60fps 유지 (16.67ms frame time 이하)
- **테마 전환 속도**: 200ms 이하
- **로그 검색 응답**: 500ms 이하
- **번들 크기 증가**: 50KB 이하

### Quality Gates
- **테스트 커버리지**: 새로운 코드 90% 이상
- **TypeScript 컴파일**: 0 에러, 0 경고
- **ESLint/Prettier**: 모든 규칙 통과
- **Lighthouse 성능**: 90점 이상 유지

### Acceptance Criteria
- 스크롤 문제 100% 해결 (모든 브라우저)
- 한글 코멘트 자동화 100% 성공률
- 테마 전환 실패율 < 0.1%
- 로그 데이터 손실 0건

## Estimated Effort

### Overall Timeline
- **Total Duration**: 3주 (15 working days)
- **Critical Path**: Dashboard Scroll Fix → Theme System → Korean Comments
- **Buffer Time**: 20% (3일) 예상치 못한 이슈 대응

### Resource Requirements  
- **Frontend Developer**: 1명 풀타임 (15일)
- **QA Engineer**: 0.5명 (마지막 1주)
- **Code Reviews**: 매일 1시간씩

### Risk Buffer
- **High Risk Items**: Dashboard scroll fix (2일 추가 버퍼)
- **Medium Risk Items**: Theme system (1일 추가 버퍼)  
- **Integration Testing**: 2일 추가 버퍼

### Critical Path Items
1. **Week 1**: Scroll fix 완료 필수 (다른 작업의 전제 조건)
2. **Week 2**: Theme system 완료 필수 (UI 통합 작업의 기반)
3. **Week 3**: 모든 기능 통합 및 테스트

---

**Total Effort Estimate**: 21 person-days (15 development + 6 buffer/QA)