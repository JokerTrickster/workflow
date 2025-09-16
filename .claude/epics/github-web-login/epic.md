---
name: github-web-login
status: backlog
created: 2025-09-06T13:00:55Z
progress: 0%
prd: .claude/prds/github-web-login.md
github: https://github.com/JokerTrickster/workflow/issues/1
---

# Epic: github-web-login

## Overview

프론트엔드 중심의 GitHub OAuth 인증 시스템으로, Next.js 15와 기존 기술 스택을 활용하여 최소한의 코드로 구현합니다. Supabase Auth를 활용한 OAuth 플로우와 GitHub REST API 직접 연동으로 레포지토리 데이터를 가져오는 간결한 아키텍처입니다.

## Architecture Decisions

- **Authentication**: Supabase Auth + GitHub OAuth (기존 스택 활용, 복잡한 토큰 관리 제거)
- **API Layer**: Next.js 15 App Router API Routes (단순 프록시 역할)
- **State Management**: @tanstack/react-query (서버 상태만 관리, 복잡한 글로벌 상태 없음)
- **UI Framework**: shadcn-ui + TailwindCSS (기존 디자인 시스템 활용)
- **Storage**: Browser LocalStorage + Supabase session (간단한 토큰 관리)

## Technical Approach

### Frontend Components
**핵심 컴포넌트 3개로 최소화:**
- `LoginPage`: GitHub OAuth 버튼과 로딩 상태만 처리
- `DashboardPage`: 레포지토리 카드 그리드 레이아웃
- `RepoCard`: 개별 레포지토리 메타데이터 표시

**상태 관리 패턴:**
- React Query로 서버 상태 (레포지토리 목록, 사용자 정보) 캐싱
- 로컬 상태 최소화 (로딩, 에러 상태만)
- Supabase 세션으로 인증 상태 관리

### Backend Services
**API 엔드포인트 (최소 2개):**
- `GET /api/auth/session`: Supabase 세션 검증 및 사용자 정보 반환
- `GET /api/github/repositories`: GitHub API 프록시 (페이지네이션 포함)

**비즈니스 로직:**
- GitHub API Rate Limit 헤더 전달
- 레포지토리 메타데이터 정규화 (이름, 설명, 언어, 별표, 업데이트 시간)
- 에러 처리 및 재시도 로직

### Infrastructure
- **Deployment**: Vercel (Next.js 최적화)
- **Environment**: GitHub OAuth App 설정 + Supabase 프로젝트
- **Monitoring**: Vercel Analytics + 기본 로깅

## Implementation Strategy

**Phase 1: 인증 기반 구축 (2일)**
- Supabase 프로젝트 설정 + GitHub OAuth Provider 연동
- LoginPage 컴포넌트 + 기본 라우팅
- 세션 관리 및 리다이렉션 플로우

**Phase 2: 레포지토리 연동 (2일)**
- GitHub API 프록시 엔드포인트
- DashboardPage + RepoCard 컴포넌트
- React Query 데이터 패칭 및 캐싱

**Phase 3: UX 개선 (1일)**
- 에러 처리 + 로딩 상태 개선
- 반응형 레이아웃 최적화
- 검색/필터 기능 추가

## Task Breakdown Preview

High-level task categories that will be created:
- [ ] **Auth Setup**: Supabase 프로젝트 + GitHub OAuth 설정
- [ ] **Login UI**: 로그인 페이지 + OAuth 플로우 구현
- [ ] **GitHub API**: 레포지토리 데이터 가져오기 API 엔드포인트
- [ ] **Dashboard UI**: 레포지토리 목록 표시 + 카드 컴포넌트
- [ ] **Session Management**: 인증 상태 관리 + 리다이렉션
- [ ] **Error Handling**: 에러 처리 + 사용자 친화적 메시지
- [ ] **Responsive Design**: 모바일/데스크톱 최적화
- [ ] **Testing & Polish**: 기본 테스트 + 성능 최적화

## Dependencies

**External Service Dependencies:**
- GitHub OAuth App 등록 (Client ID/Secret 발급)
- Supabase 프로젝트 생성 + GitHub Provider 설정
- Vercel 배포 환경 설정

**Internal Dependencies:**
- 기존 Next.js 15 프로젝트 구조 활용
- shadcn-ui 컴포넌트 라이브러리 설정 완료
- TailwindCSS 설정 및 디자인 토큰 정의

**Prerequisite Work:**
- GitHub OAuth App 생성 (5분)
- Supabase 프로젝트 초기 설정 (10분)
- 환경 변수 구성 (.env.local)

## Success Criteria (Technical)

**Performance Benchmarks:**
- 첫 페이지 로드: < 2초 (Lighthouse Score 90+)
- GitHub API 응답: < 3초 (100개 레포지토리 기준)
- OAuth 인증 완료: < 5초

**Quality Gates:**
- TypeScript 컴파일 에러 0개
- ESLint 경고 0개
- 모든 API 엔드포인트 에러 핸들링 구현
- 반응형 디자인 검증 (360px ~ 1440px)

**Acceptance Criteria:**
- GitHub 로그인 성공률 95% 이상
- 모든 레포지토리 (public + private) 표시
- 토큰 만료 시 자동 재인증
- 모바일/데스크톱 모든 환경에서 정상 작동

## Estimated Effort

**Overall Timeline**: 5일 (1주 이내 완료)
**Critical Path Items:**
1. GitHub OAuth App 설정 (Day 1)
2. Supabase Auth 연동 (Day 1-2)
3. GitHub API 연동 (Day 2-3)
4. UI 컴포넌트 구현 (Day 3-4)
5. 테스트 및 최적화 (Day 5)

**Resource Requirements:**
- 1인 개발자 (캡틴)
- GitHub 개발자 계정
- Supabase 무료 플랜
- Vercel 배포 환경

**Risk Mitigation:**
- OAuth 설정 실패 → 공식 문서 및 예제 코드 활용
- API Rate Limit → 캐싱 전략 및 요청 최적화
- UI 복잡성 → shadcn-ui 기존 컴포넌트 최대한 활용

## Tasks Created
- [ ] 001.md - Auth Setup - Supabase 프로젝트 + GitHub OAuth 설정 (parallel: false)
- [ ] 002.md - Login UI - 로그인 페이지 + OAuth 플로우 구현 (parallel: true)
- [ ] 003.md - Session Management - 인증 상태 관리 + 리다이렉션 (parallel: true)
- [ ] 004.md - GitHub API - 레포지토리 데이터 가져오기 API 엔드포인트 (parallel: true)
- [ ] 005.md - Dashboard UI - 레포지토리 목록 표시 + 카드 컴포넌트 (parallel: false)
- [ ] 006.md - Error Handling - 에러 처리 + 사용자 친화적 메시지 (parallel: false)
- [ ] 007.md - Responsive Design - 모바일/데스크톱 최적화 (parallel: true)
- [ ] 008.md - Testing & Polish - 기본 테스트 + 성능 최적화 (parallel: false)

Total tasks: 8
Parallel tasks: 4
Sequential tasks: 4
Estimated total effort: 64 hours (8 days)