---
github: "https://github.com/JokerTrickster/workflow/issues/2"
last_sync: "2025-09-26T17:16:15.860734Z"
status: completed

---

# Task: Auth Setup - Supabase 프로젝트 + GitHub OAuth 설정

프로젝트의 인증 인프라를 구축합니다. Supabase 프로젝트 생성, GitHub OAuth App 등록, 환경 설정, Next.js와의 연동을 포함하는 기반 작업입니다.

## Acceptance Criteria
- [ ] Supabase 프로젝트 생성 및 초기 설정 완료
- [ ] GitHub OAuth App 등록 및 Client ID/Secret 발급
- [ ] .env.local 환경 변수 구성 완료
- [ ] Next.js 프로젝트에 Supabase 클라이언트 설정
- [ ] GitHub OAuth Provider 연동 설정
- [ ] 로컬 개발 환경에서 인증 플로우 테스트 가능

## Technical Details
**구현할 파일 및 설정:**
- Supabase 프로젝트 대시보드 설정
- GitHub OAuth App 등록 (callback: http://localhost:3000/auth/callback)
- 환경 변수: NEXT_PUBLIC_SUPABASE_URL, NEXT_PUBLIC_SUPABASE_ANON_KEY, SUPABASE_SERVICE_ROLE_KEY
- lib/supabase.ts: Supabase 클라이언트 초기화
- GitHub OAuth Provider 설정 (repo, user:email 권한)

## Dependencies
- [ ] 없음 (기반 작업)

## Effort Estimate
- Size: S (8 hours)
- Timeline: 1 day
- Parallel: false (다른 모든 작업의 전제 조건)

## Definition of Done
- [ ] Supabase 프로젝트 생성 완료
- [ ] GitHub OAuth App 등록 완료
- [ ] 환경 변수 설정 완료
- [ ] Supabase 클라이언트 설정 완료
- [ ] 인증 플로우 기본 구성 완료
- [ ] 로컬 테스트 환경 동작 확인

## Related
- Epic: #1
- Task File: .claude/epics/github-web-login/001.md