---
name: s3-static-hosting
status: backlog
created: 2025-10-02T00:26:29Z
progress: 0%
prd: .claude/prds/s3-static-hosting.md
github: https://github.com/JokerTrickster/workflow/issues/95
---

# Epic: S3 Static Hosting Migration

## Overview

Migrate Next.js frontend from SSR to static export by eliminating API routes and enabling direct backend communication. The approach is minimal: delete 8 proxy API routes, configure static export, and leverage existing `apiClient` infrastructure that already supports direct backend calls. No new client-side code needed - the foundation is already there.

**Core Strategy**:
- Remove API route middleware layer (8 files)
- Backend implements 5 new endpoints (work logs, epics, GitHub proxy)
- Frontend already uses `apiClient` - just update base URL
- Configure CORS on backend, deploy static bundle to S3

## Architecture Decisions

### AD1: Leverage Existing ApiClient Infrastructure
**Decision**: Use existing `ApiClient.ts` without modification
**Rationale**: Frontend already has `apiClient` with `getApiBaseUrl()`, auth headers, error handling. All task operations already use it. Only work logs and epics need migration from API routes to backend.
**Impact**: Zero new frontend architecture needed, just update service URLs

### AD2: Backend Owns All File Operations
**Decision**: Move `.claude/logs` and `.claude/epics` file I/O to backend
**Rationale**: Static frontend cannot access file system. Backend already exists at `http://13.203.37.93:7000`, just add 5 endpoints.
**Impact**: Backend must implement markdown parsing (gray-matter) and file CRUD

### AD3: Static Export Without Image Optimization
**Decision**: Use `output: 'export'` with `images: { unoptimized: true }`
**Rationale**: Zero `next/image` usage in codebase (already verified). No optimization needed.
**Impact**: Simple config change, no code refactoring

### AD4: HTTP-Only S3 Hosting
**Decision**: Deploy to S3 static website hosting without CloudFront
**Rationale**: Personal use application, no HTTPS requirement specified
**Impact**: Simple bucket policy, no CDN setup, faster deployment

### AD5: CORS Middleware on Backend
**Decision**: Add CORS middleware to allow S3 origin
**Rationale**: Browser security requires CORS headers for cross-origin requests
**Impact**: Backend middleware change, must support OPTIONS preflight

## Technical Approach

### Frontend Changes (Minimal)

**Configuration** (2 files):
- `next.config.ts`: Add `output: 'export'`, `images: { unoptimized: true }`
- `.env.production`: Set `NEXT_PUBLIC_API_BASE_URL=http://13.203.37.93:7000/api/v1`

**Service Updates** (3 files):
- `WorkLogManager.ts`: Change `API_BASE = '/api/work-logs'` → `${apiClient.baseURL}/work-logs`
- `githubApi.ts`: Add `mergeBackendPullRequest()` method (calls backend proxy)
- Remove `NEXT_PUBLIC_GITHUB_TOKEN` from env (security risk)

**API Route Deletion** (8 files):
- Delete entire `/src/app/api/` directory (all routes are proxies or file ops)

### Backend Changes (New Functionality)

**CORS Middleware**:
```go
Access-Control-Allow-Origin: http://<s3-bucket-url>
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

**New Endpoints** (5 total):
1. `GET/POST /api/v1/work-logs` - Read/write work log files
2. `POST /api/v1/work-logs/entry` - Append to daily log
3. `PUT /api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge` - Proxy GitHub merge
4. `GET/POST /api/v1/epics/tasks` - Epic task list/create
5. `GET/PUT/DELETE /api/v1/epics/tasks/{id}` - Epic task CRUD

**File I/O Operations**:
- Read/write `.claude/logs/{repo}/{date}.md`
- Parse/write `.claude/epics/repositories/{repo}/tasks/{id}.md`
- Use `gray-matter` for frontmatter parsing

### Infrastructure

**S3 Configuration**:
- Bucket policy: Public read access
- Static website hosting enabled
- Index document: `index.html`
- Error document: `404.html` (for client-side routing)

**Deployment Flow**:
1. Build: `npm run build` → generates `out/` directory
2. Upload: `aws s3 sync out/ s3://bucket-name/`
3. Test: Verify from `http://bucket-name.s3-website-region.amazonaws.com`

## Implementation Strategy

### Phase 1: Backend Preparation (3-5 days)
Focus: Implement 5 new endpoints before frontend changes
- Implement work log endpoints with file I/O
- Implement epic endpoints with frontmatter parsing
- Implement GitHub merge proxy
- Add CORS middleware
- Test with Postman/curl

### Phase 2: Frontend Migration (2-3 days)
Focus: Delete API routes, update services, configure static export
- Update `next.config.ts` for static export
- Delete `/src/app/api/` directory (8 files)
- Update 3 service files (WorkLogManager, githubApi, env)
- Build and test locally with `npx serve out/`

### Phase 3: Deployment & Validation (1-2 days)
Focus: S3 deployment and production testing
- Create S3 bucket with static hosting
- Upload static bundle
- Update backend CORS with S3 origin
- Smoke test all features

## Task Breakdown Preview

**Critical Path** (must be sequential):
1. **Backend API Implementation** - 5 endpoints + CORS (blocks everything)
2. **Frontend Static Migration** - Delete API routes, update services
3. **S3 Deployment** - Upload and configure

**Parallel Opportunities**:
- Frontend config changes (can happen before backend done)
- Documentation updates (during development)

**Simplified Task List** (8 tasks total):
- [ ] **Task 1**: Implement backend work logs endpoints (GET/POST `/api/v1/work-logs`, POST `/api/v1/work-logs/entry`)
- [ ] **Task 2**: Implement backend epic tasks endpoints (GET/POST/PUT/DELETE `/api/v1/epics/tasks`)
- [ ] **Task 3**: Implement backend GitHub merge proxy (PUT `/api/v1/github/repos/.../merge`)
- [ ] **Task 4**: Configure backend CORS middleware (allow S3 origin, OPTIONS support)
- [ ] **Task 5**: Update frontend Next.js config and environment variables (static export settings)
- [ ] **Task 6**: Delete API routes and update frontend services (8 deletions + 3 updates)
- [ ] **Task 7**: Deploy to S3 with static website hosting (bucket policy, upload, configure)
- [ ] **Task 8**: End-to-end validation and smoke testing (all features working)

## Dependencies

### Blocking Dependencies
1. **Backend Implementation** (Tasks 1-4) → Blocks frontend deployment
   - Frontend API calls fail without backend endpoints
   - CORS errors prevent any cross-origin requests

2. **Frontend Migration** (Tasks 5-6) → Blocks S3 deployment
   - Cannot build static bundle with API routes present
   - Services must call backend directly before deployment

3. **S3 Bucket Creation** (Task 7) → Blocks production access
   - Need S3 URL to update backend CORS whitelist

### Non-Blocking Dependencies
- GitHub PAT (already exists on backend)
- Supabase OAuth (already configured)
- AWS S3 account (assumed available)

## Success Criteria (Technical)

### Build & Deploy
- [ ] `npm run build` completes without errors, generates `out/` directory
- [ ] `out/` contains only static files (HTML, JS, CSS, assets)
- [ ] No server-side code in production bundle (grep verification)
- [ ] S3 bucket serves application at HTTP URL

### Functional Validation
- [ ] All 8 API routes deleted, zero references remain in codebase
- [ ] Work logs create/retrieve works from S3-hosted frontend
- [ ] Epic tasks CRUD works from S3-hosted frontend
- [ ] GitHub PR merge works through backend proxy
- [ ] Task execution and status polling works
- [ ] Authentication persists (Supabase session)

### Security & Performance
- [ ] No `NEXT_PUBLIC_GITHUB_TOKEN` in client bundle (grep verification)
- [ ] No CORS errors in browser console (DevTools check)
- [ ] Direct backend calls complete in ≤500ms (95th percentile)
- [ ] Static assets cached properly (S3 headers)

### Rollback Readiness
- [ ] Previous `out/` directory backed up before deployment
- [ ] Backend CORS can be reverted to localhost-only
- [ ] Deployment procedure documented for rollback

## Estimated Effort

### Task-Level Estimates

| Task | Component | Hours | Complexity |
|------|-----------|-------|------------|
| 1 | Backend work logs | 4-6h | Medium (file I/O + markdown) |
| 2 | Backend epic tasks | 4-6h | Medium (frontmatter parsing) |
| 3 | Backend GitHub proxy | 2-3h | Low (simple proxy) |
| 4 | Backend CORS | 2-3h | Low (middleware config) |
| 5 | Frontend config | 1-2h | Low (config files) |
| 6 | Frontend migration | 3-4h | Low (deletions + URL updates) |
| 7 | S3 deployment | 2-3h | Low (bucket setup) |
| 8 | E2E validation | 3-4h | Medium (full testing) |
| **Total** | **8 tasks** | **21-31h** | **3-4 days** |

### Critical Path Duration
- Backend implementation: 12-18h (tasks 1-4)
- Frontend migration: 4-6h (tasks 5-6)
- Deployment & validation: 5-7h (tasks 7-8)
- **Minimum duration**: 21 hours (~3 days with no blockers)

### Resource Requirements
- **Developer**: 1 (solo, doing both frontend & backend)
- **Infrastructure**: AWS S3 account (assumed available)
- **External Services**: None (backend, Supabase, GitHub already configured)

## Risk Mitigation

### High-Risk Mitigations
1. **CORS Failures** → Test CORS with curl before frontend deployment
2. **Backend Delays** → Complete backend (tasks 1-4) before starting frontend
3. **File I/O Bugs** → Test work log/epic operations thoroughly with Postman

### Medium-Risk Mitigations
4. **S3 Misconfiguration** → Use AWS docs for bucket policy template
5. **Client-Side Routing** → Configure error document to `404.html` for SPA routing

### Contingency Plans
- **Backend issues**: Rollback to SSR version, keep API routes temporarily
- **CORS issues**: Add S3 origin to backend whitelist, verify OPTIONS responses
- **Deployment issues**: Redeploy previous `out/` directory from backup

## Definition of Done

Epic is complete when:
- [ ] All 8 tasks marked as completed
- [ ] Application accessible from S3 HTTP URL
- [ ] All features work identically to SSR version
- [ ] Zero CORS errors in production
- [ ] No security vulnerabilities (no PATs in bundle)
- [ ] Rollback procedure tested and documented
- [ ] User (you) confirms application works as expected

## Tasks Created

- [ ] #96 - Implement Backend Work Logs Endpoints (parallel: true)
- [ ] #98 - Implement Backend Epic Tasks Endpoints (parallel: true)
- [ ] #102 - Implement Backend GitHub Merge Proxy (parallel: true)
- [ ] #97 - Configure Backend CORS Middleware (parallel: true)
- [ ] #99 - Update Frontend Next.js Config and Environment Variables (parallel: true)
- [ ] #101 - Delete API Routes and Update Frontend Services (parallel: false)
- [ ] #100 - Deploy to S3 with Static Website Hosting (parallel: false)
- [ ] #103 - End-to-End Validation and Smoke Testing (parallel: false)

Total tasks: 8
Parallel tasks: 5
Sequential tasks: 3
Estimated total effort: 21-31 hours
