AI Git Workbench – Product Requirements Document (PRD)
1. 개요
1인 개발자가 여러 GitHub 레포지토리를 AI 에이전트(Claude + GPT-4/5)와 함께 병렬로 작업·관리할 수 있는 Web + 안드로이드 전용 워크벤치.

2. 문제 정의
여러 프로젝트를 동시에 개발할 때 레포 동기화·브랜치 관리·PR 작성 작업이 반복적이고 번거롭다.
AI 작업 흐름이 길어지면 토큰 부족으로 중단되거나 진행 이력이 누락된다.
진행 현황과 빌드 결과를 한 화면에서 파악하기 어렵다.
3. 목표
한 화면에서 레포 연결→작업 생성→자동 PR까지 완전 자동화.
반복 작업 최소화로 개인 생산성 2배 향상.
토큰 사용 대비 완료 작업 비율 70% 이상 달성.
MVP를 1개월 이내 출시하여 실제 코딩 업무에 투입.
4. 핵심 지표(KPI)
(A) 일일 자동 PR 생성 수
(B) 코드 품질 점수(빌드·Lint 통과율)
(C) 토큰 사용 대비 완료 작업 비율
5. 타깃 사용자
1인 개발자(초·중급) ➜ 현재 사용자는 ‘캡틴’ 단 한 명.
6. 주요 사용 시나리오
캡틴이 웹·앱에 로그인 → GitHub 레포 동기화
AI 에이전트에게 새로운 작업을 지시 → 자동 브랜치 생성·코드 수정
빌드 & Lint 검증 후 자동 PR + 알림 전송
여러 프로젝트를 병렬로 돌리며 진행 현황·토큰 사용량을 대시보드로 확인
7. 범위 (MVP)
7.1 필수 기능
GitHub OAuth 로그인 & 레포 목록 동기화 (없으면만 다운로드)
레포 리스트 UI + ‘연동’ 버튼 → 워크스페이스 패널(레포/작업/로그)
작업 관리
Markdown 파일 기반 작업 리스트 (.md)
작업 추가·시작·삭제
진행/미시작/종료 분류 & 정렬
AI 작업 실행
Claude + GPT-4/5 연동
새 브랜치 생성 → 코드 수정 → 빌드·Lint → PR 작성
PR 설명: 작업 내용·추가 개선점·테스트 결과
토큰 모니터링
10% 미만 시 작업 히스토리를 .md 로 저장하고 자동 중단
로그 탭: 모든 완료 작업 히스토리 출력
실시간 푸시 알림 (빌드 완료·PR 생성·토큰 부족 등)
코드 스니펫 뷰어 (모바일 최적화)
플랫폼 지원: 웹(반응형 Next.js 15) + 안드로이드(React Native Android 빌드)
7.2 부가 기능 (Nice-to-Have)
AI 작업 요약 대시보드
오프라인 모드 & 후속 동기화
팀원 초대(향후)
7.3 제외 범위(Non-Features)
GitHub 외 Git 서비스 연동
결제·구독 모델
iOS 앱
복잡한 DB 설계(초기에는 파일 시스템만)
8. 기술 요구 사항
BASE STACK: Next.js 15, TypeScript, TailwindCSS, shadcn-ui, lucide-react, @tanstack/react-query, date-fns, Supabase(SDK 사용은 로그인 정도만, DB X)
안드로이드: React Native 0.74 + Expo Bare(빌드·푸시 알림 지원)
저장소: 로컬 파일 시스템(향후 S3 옵션)
백엔드: Next.js App Router 내 API Route + node-git + simple-git
푸시 알림: Firebase FCM(Android) + Web Push
AI Gateway: Serverless 함수에서 Claude & GPT OpenAI API Key 라우팅
9. 사용자 경험(UX) 요구 사항
레포 홈 > 연결된 레포 카드 > 워크스페이스 패널 3탭(레포 / 작업 / 로그)
다크·라이트 모드 자동 대응
모바일(360 px) ↔ 데스크톱(1440 px) 레이아웃 유지
10. 성능·보안 요구 사항
초기 로드 < 2 초 (캐시된 레포 목록 기준)
GitHub PAT/OAuth 토큰 암호화 저장(Local Secure Storage)
API 오류·토큰 만료 시 즉시 재로그인 플로우
11. 릴리스 플랜 (1개월)
주차	마일스톤
1주차	프로젝트 부트스트랩, GitHub OAuth, 레포 목록 UI
2주차	작업 .md 구조, 워크스페이스 패널, 로컬 selective clone
3주차	AI 연동(Claude+GPT) → 브랜치·PR 파이프라인, 토큰 모니터링
4주차	푸시 알림, 안드로이드 빌드, 베타 테스트 & 버그픽스
12. 리스크 & 완화 전략
AI API 비용 증가 ➜ 토큰 효율 모니터, 작업 배치 크기 제한
로컬 저장소 용량 한계 ➜ 오래된 레포 캐시 삭제 설정
Android FCM 설정 복잡 ➜ Expo EAS 사용, 최소 기능부터 적용
13. 오픈 이슈
멀티-모델 라우팅 로직 상세 설계 필요
PR 자동 병합 규칙 정의(리뷰 스킵 여부)
장기적으로 DB 도입 시 마이그레이션 전략
