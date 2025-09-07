---
name: github-web-login
description: GitHub OAuth 연동 로그인 시스템 및 레포지토리 정보 가져오기 기능 구현
status: backlog
created: 2025-09-06T12:50:31Z
---

# PRD: GitHub Web Login

## Executive Summary

AI Git Workbench 프로젝트의 핵심 인증 시스템인 GitHub OAuth 로그인 기능을 구현합니다. 사용자가 GitHub 계정으로 로그인하고, 인증 성공 시 메인 페이지로 이동하여 사용자의 모든 GitHub 레포지토리 정보를 가져오는 프론트엔드 중심의 기능입니다.

**핵심 가치**: 원클릭 GitHub 연동을 통해 개발자의 워크플로우 시작점을 간소화하고, 즉시 모든 레포지토리에 접근 가능한 대시보드 제공

## Problem Statement

### 해결하고자 하는 문제
- 현재 AI Git Workbench에 인증 시스템이 없어 GitHub 레포지토리에 접근할 수 없음
- 개발자가 여러 레포지토리를 수동으로 찾아 연동해야 하는 번거로움
- GitHub API 접근 권한 없이는 레포지토리 정보 및 작업 수행이 불가능

### 왜 지금 중요한가?
- AI Git Workbench MVP의 첫 번째 필수 기능 (1주차 마일스톤)
- 후속 기능들(작업 관리, AI 연동, PR 생성)의 선행 조건
- 사용자가 앱을 처음 실행했을 때의 첫인상을 결정하는 핵심 기능

## User Stories

### 주요 사용자 페르소나
**캡틴**: 1인 개발자(초·중급), 여러 GitHub 레포지토리를 동시에 관리하는 개발자

### 상세 사용자 여정

#### Story 1: 첫 방문 및 로그인
**As a** 개발자  
**I want to** GitHub 계정으로 앱에 로그인하고 싶다  
**So that** 내 레포지토리들에 접근할 수 있다

**Acceptance Criteria:**
- 앱 첫 실행 시 GitHub 로그인 버튼이 명확하게 표시됨
- GitHub OAuth 인증 플로우가 새 탭/팝업에서 열림
- 사용자가 GitHub에서 앱 권한을 승인할 수 있음
- 승인 후 자동으로 메인 페이지로 리디렉션됨

#### Story 2: 레포지토리 목록 확인
**As a** 로그인한 사용자  
**I want to** 내 모든 GitHub 레포지토리를 한 눈에 보고 싶다  
**So that** 작업할 레포지토리를 선택할 수 있다

**Acceptance Criteria:**
- 로그인 성공 시 메인 페이지에서 모든 레포지토리 목록이 카드 형태로 표시됨
- 각 레포지토리 카드에는 이름, 설명, 언어, 별표 수, 마지막 업데이트 시간이 표시됨
- Public/Private 레포지토리 구분이 명확함
- 레포지토리 검색 및 필터링 기능 제공

#### Story 3: 오류 처리 및 재인증
**As a** 사용자  
**I want to** 로그인 오류나 토큰 만료 시 명확한 안내를 받고 싶다  
**So that** 문제를 해결하고 계속 사용할 수 있다

**Acceptance Criteria:**
- GitHub OAuth 인증 실패 시 사용자 친화적인 오류 메시지 표시
- 토큰 만료 시 자동으로 재로그인 플로우 실행
- 네트워크 오류 시 재시도 옵션 제공

## Requirements

### Functional Requirements

#### 1. GitHub OAuth 인증
- GitHub OAuth App 설정 및 연동
- Authorization Code 플로우 구현
- Access Token 획득 및 저장
- 사용자 프로필 정보 가져오기

#### 2. 레포지토리 데이터 가져오기
- GitHub REST API v4 연동
- 사용자의 모든 레포지토리 목록 조회 (public + private)
- 레포지토리 메타데이터 표시 (이름, 설명, 언어, 별표, 포크, 업데이트 시간)
- Pagination 처리 (100개 이상 레포지토리 지원)

#### 3. 사용자 인터페이스
- 반응형 로그인 페이지 (모바일 360px ~ 데스크톱 1440px)
- GitHub 브랜딩 가이드라인 준수한 로그인 버튼
- 레포지토리 카드 레이아웃 (그리드 형태)
- 다크/라이트 모드 지원

#### 4. 세션 관리
- Access Token 보안 저장 (Local Secure Storage)
- 자동 로그인 유지 (Remember Me)
- 안전한 로그아웃 기능

### Non-Functional Requirements

#### 1. 성능
- 초기 로그인 페이지 로딩 시간: 2초 이내
- GitHub API 응답 대기 시간: 5초 이내
- 레포지토리 목록 표시: 3초 이내

#### 2. 보안
- OAuth 2.0 표준 준수
- GitHub Personal Access Token 암호화 저장
- HTTPS 통신 강제
- XSS, CSRF 공격 방지

#### 3. 호환성
- Next.js 15 환경에서 작동
- 주요 브라우저 지원 (Chrome, Firefox, Safari, Edge)
- 모바일 브라우저 최적화

#### 4. 사용성
- 직관적인 UI/UX (GitHub 사용자에게 친숙함)
- 접근성 표준 준수 (WCAG 2.1 AA)
- 다국어 지원 준비 (한국어 우선)

## Success Criteria

### 핵심 성공 지표
1. **로그인 성공률**: 95% 이상
2. **초기 로딩 시간**: 2초 이내
3. **API 오류율**: 5% 이하
4. **사용자 이탈률**: 로그인 페이지에서 10% 이하

### 정량적 목표
- GitHub OAuth 인증 성공률: 95%
- 레포지토리 데이터 로딩 성공률: 98%
- 페이지 로딩 속도: Lighthouse 점수 90점 이상
- 토큰 보안 저장 성공률: 100%

### 정성적 목표
- 사용자가 직관적으로 로그인할 수 있음
- GitHub 브랜딩과 일관된 디자인
- 오류 상황에서 명확한 가이드 제공

## Constraints & Assumptions

### 기술적 제약사항
- Next.js 15 프레임워크 사용 필수
- GitHub OAuth Apps 정책 준수
- GitHub API Rate Limit (5000 requests/hour) 고려
- 로컬 파일 시스템 기반 저장소 사용

### 비즈니스 제약사항
- 1인 개발자 (캡틴) 전용 (초기 MVP)
- 무료 GitHub API 사용
- 1주차 내 구현 완료 필요

### 기술적 가정
- GitHub API가 안정적으로 작동
- 사용자의 브라우저가 모던 JavaScript를 지원
- 인터넷 연결 상태 양호
- Local Storage 사용 가능

## Out of Scope

### 명시적으로 구현하지 않을 기능
- 다른 Git 서비스 연동 (GitLab, Bitbucket)
- 소셜 로그인 (Google, Facebook)
- 사용자 등록/회원가입 시스템
- 비밀번호 기반 인증
- 팀/조직 계정 관리
- 결제/구독 시스템
- iOS 앱 지원

### 차후 단계로 미룰 기능
- Advanced 권한 관리
- GitHub Enterprise Server 지원
- 멀티 계정 로그인
- SSO 연동

## Dependencies

### External Dependencies
- **GitHub OAuth Apps**: 앱 등록 및 Client ID/Secret 발급 필요
- **GitHub REST API v4**: 레포지토리 데이터 조회
- **Next.js 15**: 프론트엔드 프레임워크
- **Supabase**: OAuth 인증 처리 지원

### Internal Dependencies
- **UI 컴포넌트**: shadcn-ui 기반 Button, Card 컴포넌트
- **상태 관리**: @tanstack/react-query로 API 상태 관리
- **스타일링**: TailwindCSS 설정 완료

### Team Dependencies
- **캡틴**: GitHub OAuth App 설정 및 환경 변수 구성
- **디자인 가이드**: GitHub 브랜딩 가이드라인 참조

## Implementation Notes

### 기술 스택 확인
기존 TRD에서 정의된 스택 활용:
- Next.js 15 (App Router)
- TypeScript
- TailwindCSS
- shadcn-ui
- @tanstack/react-query
- supabase-js

### 파일 구조 (TRD 기준)
```
pages/
├── api/
│   └── auth/
│       └── github.ts          # GitHub OAuth 콜백 처리
├── login.tsx                  # 로그인 페이지
└── dashboard.tsx              # 메인 대시보드 (레포 목록)

components/
├── LoginButton.tsx            # GitHub 로그인 버튼
├── RepoCard.tsx              # 레포지토리 카드
└── RepoList.tsx              # 레포지토리 목록

utils/
├── auth.ts                   # 인증 헬퍼 함수
└── api.ts                    # API 호출 헬퍼

services/
└── github.ts                 # GitHub API 연동
```

### 환경 설정 필요사항
```
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your_secret_key
```