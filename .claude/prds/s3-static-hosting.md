---
name: s3-static-hosting
description: Migrate Next.js frontend from SSR to S3 static hosting with direct backend API calls
status: backlog
created: 2025-10-02T00:19:41Z
---

# PRD: S3 Static Hosting Migration

## Executive Summary

This project migrates the Next.js 15.5.2 frontend application from Server-Side Rendering (SSR) to fully static hosting on AWS S3. The current architecture uses 8 Next.js API routes as proxies to backend services, which prevents static deployment. The migration will eliminate all API routes, configure direct browser-to-backend communication, implement required CORS policies, and generate a static HTML/JS/CSS bundle deployable to S3.

**Key Objectives:**
- Remove all 8 Next.js API routes and replace with direct backend calls
- Configure Next.js for static export (`output: 'export'`)
- Implement CORS on backend to accept browser requests from S3
- Migrate server-side file operations (work logs, epic tasks) to backend APIs
- Deploy static bundle to S3 with HTTP-only hosting (no CloudFront)
- Maintain 100% feature parity with current SSR implementation

**Business Value:**
- Simplified deployment (no server runtime required)
- Reduced infrastructure complexity (static files vs Node.js server)
- Improved performance (direct backend calls, no proxy overhead)
- Cost reduction (S3 storage vs compute instances)

## Problem Statement

### Current Architecture Limitations

The existing Next.js application cannot be deployed to S3 static hosting because:

1. **API Routes Dependency**: 8 server-side API routes require Node.js runtime
   - `/api/v1/tasks/*` - Backend task proxies (3 routes)
   - `/api/work-logs/*` - File system operations (2 routes)
   - `/api/github/*/merge` - GitHub API proxy (1 route)
   - `/api/epics/tasks/*` - File system + backend hybrid (2 routes)

2. **File System Operations**: API routes directly access `.claude/logs` and `.claude/epics` directories on server
   - Cannot run in browser (no file system access)
   - Requires backend migration to persist data

3. **CORS Not Configured**: Backend at `http://localhost:7000` only accepts same-origin requests
   - Browser requests from S3 domain would fail
   - No CORS headers configured for cross-origin access

4. **Environment Variables**: Mix of build-time and runtime variables
   - Server-only secrets exposed in some API routes
   - Need proper categorization for static builds

### Why This Matters Now

- **User Requirement**: Personal use application, no need for SSR complexity
- **Deployment Simplicity**: S3 static hosting easier to manage than Node.js server
- **Architecture Alignment**: Backend already exists at `http://13.203.37.93:7000`, frontend should consume it directly
- **Security**: Current `NEXT_PUBLIC_GITHUB_TOKEN` exposes PAT in client bundle (should be backend-only)

## User Stories

### Primary User Persona: Solo Developer (You)

**Background:**
- Manages workflow automation through web interface
- Needs to deploy frontend without server infrastructure
- Requires all current features (tasks, work logs, GitHub integration, epic management)

### User Story 1: Direct Backend Access
**As a** user of the workflow application
**I want** the frontend to call backend APIs directly from the browser
**So that** I can deploy static files to S3 without a Node.js server

**Acceptance Criteria:**
- [ ] All API calls go directly to `http://13.203.37.93:7000` from browser
- [ ] No Next.js API routes exist in codebase
- [ ] CORS headers allow S3 origin to access backend
- [ ] All features work identically to SSR version

**Technical Notes:**
- Update `NEXT_PUBLIC_API_BASE_URL` to `http://13.203.37.93:7000/api/v1`
- Configure CORS on backend with S3 bucket origin
- Test with browser DevTools network tab

### User Story 2: Work Log Management
**As a** developer tracking work progress
**I want** to create and view work logs from the static frontend
**So that** I can maintain project history without server-side file operations

**Acceptance Criteria:**
- [ ] Work logs persist to backend storage (not local file system)
- [ ] Backend provides GET/POST `/api/v1/work-logs` endpoints
- [ ] Backend handles markdown file creation in `.claude/logs` directory
- [ ] Frontend `WorkLogManager` service updated to call backend directly
- [ ] No functionality lost from current implementation

**Technical Notes:**
- Backend must implement file I/O operations previously in `/api/work-logs/*`
- Migrate validation logic (path traversal protection) to backend
- Test log creation and retrieval end-to-end

### User Story 3: GitHub PR Operations
**As a** developer managing pull requests
**I want** to merge PRs through the web interface
**So that** I can complete workflows without exposing GitHub tokens in frontend

**Acceptance Criteria:**
- [ ] PR merge operations call backend proxy (not GitHub API directly)
- [ ] Backend implements PUT `/api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge`
- [ ] GitHub Personal Access Token stored on backend only (never exposed to browser)
- [ ] `NEXT_PUBLIC_GITHUB_TOKEN` removed from frontend environment variables
- [ ] PR merge functionality works identically to current implementation

**Technical Notes:**
- Backend uses server-side `GITHUB_TOKEN` for GitHub API calls
- Frontend `githubApi.ts` updated with `mergeBackendPullRequest()` method
- Remove client-side GitHub merge logic

### User Story 4: Epic Task Management
**As a** developer organizing work into epics
**I want** to create and manage epic tasks from the static frontend
**So that** I can maintain project structure without server-side file parsing

**Acceptance Criteria:**
- [ ] Epic tasks persist to backend storage (not local file system)
- [ ] Backend provides `/api/v1/epics/tasks` endpoints (GET/POST/PUT/DELETE)
- [ ] Backend handles markdown file parsing with frontmatter (gray-matter)
- [ ] Frontend `TaskFileManager` updated to call backend directly
- [ ] Task metadata (status, priority, assignee) properly synchronized

**Technical Notes:**
- Backend must parse YAML frontmatter and markdown content
- Migrate epic file creation logic to backend
- Test epic CRUD operations end-to-end

### User Story 5: Static Deployment
**As a** developer deploying the application
**I want** to build a static bundle and upload to S3
**So that** I can host the application without managing servers

**Acceptance Criteria:**
- [ ] `npm run build` generates static files in `out/` directory
- [ ] No server-only code remains in bundle (no API routes)
- [ ] Environment variables properly embedded at build time
- [ ] S3 bucket configured for static website hosting (HTTP only)
- [ ] Application fully functional when served from S3
- [ ] All client-side routing works with S3 index/error documents

**Technical Notes:**
- Update `next.config.ts` with `output: 'export'`
- Configure S3 bucket policy for public read access
- Set index document to `index.html`, error document to `404.html`

## Requirements

### Functional Requirements

#### FR1: API Route Elimination
- **FR1.1**: Remove all 8 Next.js API routes from `/src/app/api/` directory
- **FR1.2**: Delete `/api/v1/tasks/*` routes (3 files) - backend already implements these endpoints
- **FR1.3**: Delete `/api/work-logs/*` routes (2 files) - migrate to backend
- **FR1.4**: Delete `/api/github/*/merge` route (1 file) - migrate to backend
- **FR1.5**: Delete `/api/epics/tasks/*` routes (2 files) - migrate to backend

#### FR2: Backend API Implementation (Backend Team)
- **FR2.1**: Implement GET/POST `/api/v1/work-logs` for work log operations
- **FR2.2**: Implement POST `/api/v1/work-logs/entry` for log entry creation
- **FR2.3**: Implement PUT `/api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge` for PR merging
- **FR2.4**: Implement GET/POST `/api/v1/epics/tasks` for epic task operations
- **FR2.5**: Implement GET/PUT/DELETE `/api/v1/epics/tasks/{id}` for individual epic tasks
- **FR2.6**: Migrate file I/O operations (`.claude/logs`, `.claude/epics`) to backend
- **FR2.7**: Migrate validation logic (path traversal, repository names) to backend

#### FR3: Frontend Service Updates
- **FR3.1**: Update `WorkLogManager.ts` to call backend `/api/v1/work-logs` instead of `/api/work-logs`
- **FR3.2**: Update `githubApi.ts` to call backend `/api/v1/github/.../merge` instead of GitHub API directly
- **FR3.3**: Update `TaskFileManager.ts` to call backend `/api/v1/epics/tasks` (if exists)
- **FR3.4**: Verify `claudeService.ts` already uses `apiClient` (no changes needed)
- **FR3.5**: Verify `ActivityLogger.ts` has no API dependencies (no changes needed)

#### FR4: Environment Configuration
- **FR4.1**: Update `NEXT_PUBLIC_API_BASE_URL` from `http://localhost:7000/api/v1` to `http://13.203.37.93:7000/api/v1`
- **FR4.2**: Remove `NEXT_PUBLIC_GITHUB_TOKEN` from frontend (security risk)
- **FR4.3**: Create `.env.production` with production backend URL
- **FR4.4**: Ensure all client-side env vars have `NEXT_PUBLIC_` prefix
- **FR4.5**: Remove server-only variables (`SUPABASE_SERVICE_ROLE_KEY`, `GITHUB_TOKEN`, `REPOSITORY_NAME`)

#### FR5: Next.js Static Export
- **FR5.1**: Update `next.config.ts` with `output: 'export'`
- **FR5.2**: Configure `images: { unoptimized: true }` (no Image Optimization API in static export)
- **FR5.3**: Enable `trailingSlash: true` for S3 compatibility
- **FR5.4**: Verify no dynamic server-side rendering (no `getServerSideProps`)
- **FR5.5**: Generate static HTML for all routes at build time

#### FR6: S3 Deployment
- **FR6.1**: Create S3 bucket with static website hosting enabled (HTTP only, no CloudFront)
- **FR6.2**: Configure bucket policy for public read access
- **FR6.3**: Upload `out/` directory contents to S3 root
- **FR6.4**: Set index document to `index.html`
- **FR6.5**: Set error document to `404.html` (for client-side routing fallback)
- **FR6.6**: Verify S3 bucket URL serves application correctly

### Non-Functional Requirements

#### NFR1: Performance
- **NFR1.1**: Direct backend calls must have ≤100ms additional latency vs API route proxies
- **NFR1.2**: Static assets served from S3 with optimal caching headers
- **NFR1.3**: Bundle size unchanged or reduced (no API route code)
- **NFR1.4**: First Contentful Paint (FCP) ≤1.5s on 3G network
- **NFR1.5**: Time to Interactive (TTI) ≤3s on 3G network

#### NFR2: Security
- **NFR2.1**: No Personal Access Tokens (PATs) exposed in client bundle
- **NFR2.2**: `SUPABASE_SERVICE_ROLE_KEY` never exposed to frontend
- **NFR2.3**: CORS configured to whitelist only authorized origins (S3 bucket URL)
- **NFR2.4**: GitHub operations proxied through backend (no client-side GitHub API with PAT)
- **NFR2.5**: All API requests use HTTPS in production (note: current requirement is HTTP only)

#### NFR3: Reliability
- **NFR3.1**: CORS preflight requests handled correctly (OPTIONS method)
- **NFR3.2**: Network errors gracefully handled with user feedback
- **NFR3.3**: Backend downtime shows appropriate error messages
- **NFR3.4**: No silent failures (all errors logged to console)
- **NFR3.5**: Retry logic for transient network failures (3 retries with exponential backoff)

#### NFR4: Maintainability
- **NFR4.1**: Clear separation between frontend (S3) and backend (API server)
- **NFR4.2**: Environment variables properly documented in `.env.example`
- **NFR4.3**: Build process fully automated (`npm run build` produces deployable artifacts)
- **NFR4.4**: No manual configuration steps required for deployment
- **NFR4.5**: Rollback strategy documented (redeploy previous `out/` directory)

#### NFR5: CORS Configuration
- **NFR5.1**: Backend responds with `Access-Control-Allow-Origin: <S3-bucket-URL>`
- **NFR5.2**: Backend supports preflight OPTIONS requests for all endpoints
- **NFR5.3**: Allowed methods: GET, POST, PUT, DELETE, OPTIONS
- **NFR5.4**: Allowed headers: Content-Type, Authorization, X-Requested-With
- **NFR5.5**: `Access-Control-Allow-Credentials: true` for authenticated requests
- **NFR5.6**: `Access-Control-Max-Age: 86400` to cache preflight responses

## Success Criteria

### Primary Success Metrics

1. **Static Build Success**
   - `npm run build` generates static files in `out/` directory
   - No server-side code in production bundle
   - Bundle size ≤5MB (current baseline)

2. **Feature Parity**
   - All 8 removed API routes replaced with backend equivalents
   - Work logs create/retrieve functionality works
   - GitHub PR merge operations work
   - Epic task management works
   - Task execution and status polling works

3. **CORS Validation**
   - Zero CORS errors in browser console
   - All API requests succeed from S3 origin
   - Preflight requests complete within 100ms

4. **Security Compliance**
   - No GitHub PATs in client bundle (verify with `grep -r "ghp_" out/`)
   - No service role keys in client bundle
   - All secrets backend-only

5. **Deployment Success**
   - Application loads from S3 bucket URL
   - All routes accessible (client-side routing works)
   - All features functional end-to-end

### Key Performance Indicators (KPIs)

- **Migration Completion**: 14 files modified successfully
- **API Route Removal**: 8 API routes deleted, 0 remaining
- **Backend Endpoints**: 5 new endpoints implemented and tested
- **CORS Errors**: 0 errors in production
- **Build Time**: Static build completes in ≤60 seconds
- **Deployment Time**: S3 upload completes in ≤120 seconds

## Constraints & Assumptions

### Constraints

1. **Technical Constraints**
   - Backend must be at `http://13.203.37.93:7000` (cannot change)
   - S3 hosting must use HTTP only (no HTTPS/CloudFront requirement)
   - No server runtime available (pure static hosting)
   - CORS must be configured on backend (frontend cannot bypass)

2. **Resource Constraints**
   - Solo developer (you) implementing both frontend and backend changes
   - No dedicated QA team (self-testing required)
   - No staging environment (test locally before production)

3. **Timeline Constraints**
   - No specific deadline (flexible timeline)
   - Can be done incrementally (phase by phase)

### Assumptions

1. **Backend Availability**
   - Backend at `http://13.203.37.93:7000` is stable and accessible
   - Backend team (you) can implement required endpoints within reasonable timeframe
   - Backend supports same data models as current API routes

2. **Infrastructure**
   - AWS S3 account available for deployment
   - S3 bucket can be created without approval process
   - No AWS region restrictions (any region acceptable)

3. **Security**
   - Exposing Supabase anon key in client bundle is acceptable (intended use)
   - GitHub operations can be proxied through backend (token on backend only)
   - HTTP (not HTTPS) is acceptable for personal use application

4. **User Base**
   - Single user (you) - no multi-tenant concerns
   - No performance requirements for high traffic
   - No geographic distribution needs (single region S3)

## Out of Scope

The following items are explicitly **NOT** included in this migration:

### Not Migrating

1. **HTTPS/SSL Configuration**
   - No CloudFront setup
   - No SSL certificate provisioning
   - No HTTPS enforcement (HTTP only)

2. **Advanced S3 Features**
   - No S3 versioning
   - No lifecycle policies
   - No CDN/edge caching
   - No CloudFront distribution

3. **Server-Side Rendering**
   - No SSR/ISR support (fully static only)
   - No dynamic routes with getServerSideProps
   - No runtime page generation

4. **Image Optimization**
   - No Next.js Image Optimization API
   - No automatic image resizing/compression
   - Images served as-is (already no `next/image` usage)

5. **Advanced Security**
   - No Web Application Firewall (WAF)
   - No DDoS protection (beyond AWS defaults)
   - No rate limiting (backend responsibility)
   - No Content Security Policy (CSP) headers

6. **Multi-Environment Support**
   - No separate staging environment
   - No blue-green deployment
   - No canary releases

7. **Monitoring & Observability**
   - No application performance monitoring (APM)
   - No error tracking service (Sentry, etc.)
   - No analytics (Google Analytics, etc.)

8. **Testing Automation**
   - No E2E test suite for migration validation
   - No automated CORS testing
   - No load testing

## Dependencies

### Internal Dependencies

1. **Backend API Implementation** (Blocking)
   - **Owner**: Backend team (you)
   - **Timeline**: Required before frontend deployment
   - **Endpoints Needed**:
     - GET/POST `/api/v1/work-logs`
     - POST `/api/v1/work-logs/entry`
     - PUT `/api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge`
     - GET/POST/PUT/DELETE `/api/v1/epics/tasks` and `/api/v1/epics/tasks/{id}`
   - **Risk**: High - frontend unusable without these endpoints

2. **CORS Configuration** (Blocking)
   - **Owner**: Backend team (you)
   - **Timeline**: Required before frontend deployment
   - **Requirements**:
     - Whitelist S3 bucket origin
     - Support OPTIONS preflight requests
     - Configure allowed methods/headers
   - **Risk**: High - all API calls fail without CORS

3. **Backend File System Access** (Blocking)
   - **Owner**: Backend team (you)
   - **Timeline**: Required before work logs/epics work
   - **Requirements**:
     - Read/write access to `.claude/logs` directory
     - Read/write access to `.claude/epics` directory
     - Markdown file parsing (gray-matter library)
   - **Risk**: Medium - work logs and epics broken without file I/O

### External Dependencies

1. **AWS S3 Service** (Blocking)
   - **Provider**: Amazon Web Services
   - **Timeline**: Required for deployment
   - **Requirements**:
     - S3 bucket creation permissions
     - Public read access configuration
     - Static website hosting feature
   - **Risk**: Low - S3 is mature and stable service

2. **Supabase OAuth** (Non-blocking)
   - **Provider**: Supabase
   - **Timeline**: Already configured
   - **Requirements**: GitHub OAuth provider enabled
   - **Risk**: Low - already working in current implementation

3. **GitHub API** (Non-blocking)
   - **Provider**: GitHub
   - **Timeline**: Already configured
   - **Requirements**: PAT with repo permissions (backend-only)
   - **Risk**: Low - already working in current implementation

4. **DNS/Domain** (Optional)
   - **Provider**: DNS provider (if custom domain needed)
   - **Timeline**: Post-deployment (optional)
   - **Requirements**: CNAME record pointing to S3 bucket URL
   - **Risk**: Low - not required for basic deployment

### Technical Dependencies

1. **Next.js 15.5.2** (Current)
   - Static export support (`output: 'export'`)
   - App Router compatibility
   - No breaking changes expected

2. **React 19.1.0** (Current)
   - Client-side rendering support
   - No SSR dependencies
   - Already compatible

3. **Node.js Build Environment** (Current)
   - Node.js ≥18.x for builds
   - npm for package management
   - Already configured

## Implementation Phases

### Phase 1: Backend Preparation (Week 1)
**Owner**: Backend team
**Duration**: 3-5 days
**Dependencies**: None

**Tasks**:
1. Implement `/api/v1/work-logs` GET/POST endpoints
2. Implement `/api/v1/work-logs/entry` POST endpoint
3. Implement `/api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge` PUT endpoint
4. Implement `/api/v1/epics/tasks` CRUD endpoints
5. Configure CORS middleware with S3 origin
6. Test all endpoints with CORS preflight requests
7. Deploy backend with new endpoints to `http://13.203.37.93:7000`

**Deliverables**:
- 5 new backend endpoints operational
- CORS headers configured correctly
- Backend deployed and accessible

**Success Criteria**:
- All endpoints return 200 OK with valid responses
- OPTIONS preflight requests succeed
- Postman/curl tests pass for all endpoints

### Phase 2: Frontend Configuration (Week 1)
**Owner**: Frontend team
**Duration**: 1-2 days
**Dependencies**: None (can run parallel to Phase 1)

**Tasks**:
1. Update `next.config.ts` with `output: 'export'` and `images.unoptimized: true`
2. Create `.env.production` with production backend URL
3. Remove `NEXT_PUBLIC_GITHUB_TOKEN` from environment variables
4. Update `.env.example` with documentation

**Deliverables**:
- `next.config.ts` configured for static export
- `.env.production` with correct backend URL
- Environment variables properly categorized

**Success Criteria**:
- `npm run build` generates `out/` directory
- No build errors or warnings
- Static files properly generated

### Phase 3: API Route Removal (Week 2)
**Owner**: Frontend team
**Duration**: 1 day
**Dependencies**: Phase 1 complete (backend endpoints ready)

**Tasks**:
1. Delete `/src/app/api/v1/tasks/route.ts`
2. Delete `/src/app/api/v1/tasks/[id]/route.ts`
3. Delete `/src/app/api/v1/tasks/[id]/execute/route.ts`
4. Delete `/src/app/api/work-logs/route.ts`
5. Delete `/src/app/api/work-logs/entry/route.ts`
6. Delete `/src/app/api/github/repos/[owner]/[repo]/pulls/[number]/merge/route.ts`
7. Delete `/src/app/api/epics/tasks/route.ts`
8. Delete `/src/app/api/epics/tasks/[id]/route.ts`

**Deliverables**:
- All 8 API routes removed
- `/src/app/api/` directory empty (or deleted)

**Success Criteria**:
- No API route files remain in codebase
- `grep -r "export async function GET" src/app/api` returns empty

### Phase 4: Service Updates (Week 2)
**Owner**: Frontend team
**Duration**: 2-3 days
**Dependencies**: Phase 3 complete

**Tasks**:
1. Update `WorkLogManager.ts`: Change `API_BASE = '/api/work-logs'` to backend URL
2. Update `githubApi.ts`: Add `mergeBackendPullRequest()` method
3. Update `TaskFileManager.ts`: Change epic API calls to backend (if exists)
4. Verify `claudeService.ts` uses `apiClient` (no changes)
5. Verify `ActivityLogger.ts` has no API dependencies (no changes)

**Deliverables**:
- 3 service files updated
- All services call backend directly
- No API route references remain

**Success Criteria**:
- `grep -r "/api/" src/services` returns zero matches
- All services use `apiClient.baseURL` or `NEXT_PUBLIC_API_BASE_URL`

### Phase 5: Local Testing (Week 2)
**Owner**: Frontend team
**Duration**: 2-3 days
**Dependencies**: Phase 4 complete

**Tasks**:
1. Run `npm run build` to generate static export
2. Verify `out/` directory contains only static assets
3. Test locally with `npx serve out -p 3000`
4. Test CORS with browser DevTools network tab
5. Verify all API calls reach backend at `http://13.203.37.93:7000`
6. Test work log creation/retrieval
7. Test epic task operations
8. Test GitHub PR merge operations
9. Test task execution and status polling

**Deliverables**:
- Local static build fully functional
- All features tested and working
- CORS verified with DevTools

**Success Criteria**:
- Zero CORS errors in browser console
- All API calls return 200 OK
- All features work identically to SSR version

### Phase 6: S3 Deployment (Week 3)
**Owner**: DevOps/Frontend team
**Duration**: 1-2 days
**Dependencies**: Phase 5 complete

**Tasks**:
1. Create S3 bucket with static website hosting enabled
2. Configure bucket policy for public read access
3. Upload `out/` directory contents to S3 root
4. Update backend CORS to whitelist S3 bucket URL
5. Test application from S3 URL
6. Verify all features work with production backend

**Deliverables**:
- S3 bucket configured and deployed
- Application accessible from S3 URL
- Backend CORS updated for S3 origin

**Success Criteria**:
- Application loads from S3 bucket URL
- All routes accessible (client-side routing works)
- All API calls succeed from S3 origin

### Phase 7: Production Validation (Week 3)
**Owner**: QA/Frontend team
**Duration**: 1-2 days
**Dependencies**: Phase 6 complete

**Tasks**:
1. Smoke test: Create task and verify backend execution
2. Smoke test: Create work log entry and verify file creation
3. Smoke test: Merge PR through backend proxy
4. Monitor browser console for CORS errors
5. Monitor backend logs for API errors
6. Load test: Verify performance with direct backend calls
7. Document rollback procedure

**Deliverables**:
- Production validation report
- Performance metrics captured
- Rollback procedure documented

**Success Criteria**:
- All smoke tests pass
- Zero CORS errors in production
- Performance meets NFR targets (FCP ≤1.5s, TTI ≤3s)

## Risk Assessment

### High-Risk Items

1. **CORS Misconfiguration** (Probability: Medium, Impact: High)
   - **Risk**: Backend CORS headers incorrect, all API calls fail
   - **Mitigation**: Test CORS with curl/Postman before frontend deployment
   - **Contingency**: Rollback to SSR version, debug CORS headers

2. **Backend Endpoint Delays** (Probability: Low, Impact: High)
   - **Risk**: Backend endpoints not ready, frontend deployment blocked
   - **Mitigation**: Complete backend work in Phase 1 before frontend changes
   - **Contingency**: Deploy backend endpoints incrementally, test each

3. **File System Migration Issues** (Probability: Medium, Impact: High)
   - **Risk**: Backend file I/O implementation differs from current behavior
   - **Mitigation**: Test work log/epic operations thoroughly before deployment
   - **Contingency**: Keep API route code in separate branch for reference

### Medium-Risk Items

4. **Environment Variable Exposure** (Probability: Low, Impact: Medium)
   - **Risk**: Server-only secrets accidentally exposed in client bundle
   - **Mitigation**: Audit `out/` directory with `grep -r` for secrets before deployment
   - **Contingency**: Regenerate secrets, redeploy with correct env vars

5. **S3 Bucket Configuration** (Probability: Low, Impact: Medium)
   - **Risk**: S3 bucket policy incorrect, files not accessible
   - **Mitigation**: Use AWS documentation for bucket policy template
   - **Contingency**: Update bucket policy, clear CloudFront cache (if using)

6. **Client-Side Routing Breaks** (Probability: Low, Impact: Medium)
   - **Risk**: S3 doesn't serve index.html for nested routes
   - **Mitigation**: Configure error document to index.html for SPA routing
   - **Contingency**: Use HashRouter as fallback (not ideal)

### Low-Risk Items

7. **Bundle Size Increase** (Probability: Low, Impact: Low)
   - **Risk**: Static bundle larger than expected
   - **Mitigation**: Monitor bundle size during builds, optimize if needed
   - **Contingency**: Implement code splitting, lazy loading

8. **Build Time Increase** (Probability: Low, Impact: Low)
   - **Risk**: Static export takes longer than SSR build
   - **Mitigation**: Profile build process, optimize slow steps
   - **Contingency**: Accept longer build time (not critical for personal use)

## Rollback Strategy

### Immediate Rollback (< 1 hour)

If critical issues discovered in production:

1. **Revert S3 Bucket**:
   - Delete new static files from S3
   - Upload previous working version from backup

2. **Revert Backend CORS**:
   - Remove S3 origin from CORS whitelist
   - Keep only localhost origin

3. **Restore SSR Deployment**:
   - Redeploy previous Next.js SSR version
   - Point domain back to SSR server

### Partial Rollback (Backend Only)

If backend endpoints have issues:

1. **Disable New Endpoints**:
   - Return 503 Service Unavailable for `/api/v1/work-logs`, `/api/v1/epics`, `/api/v1/github`
   - Keep existing task endpoints operational

2. **Frontend Fallback**:
   - Display error messages for work logs/epics/GitHub features
   - Core task functionality remains working

### Full Rollback (Complete Migration Reversal)

If migration fundamentally broken:

1. **Restore API Routes**:
   - Git checkout previous commit with API routes
   - Rebuild Next.js with SSR

2. **Restore Environment Variables**:
   - Revert to `http://localhost:7000` backend URL
   - Restore server-only environment variables

3. **Redeploy SSR**:
   - Deploy to Node.js server (not S3)
   - Update DNS to point to server

## Appendix

### File Modification Checklist

**Files to Delete (8)**:
- [ ] `/src/app/api/v1/tasks/route.ts`
- [ ] `/src/app/api/v1/tasks/[id]/route.ts`
- [ ] `/src/app/api/v1/tasks/[id]/execute/route.ts`
- [ ] `/src/app/api/work-logs/route.ts`
- [ ] `/src/app/api/work-logs/entry/route.ts`
- [ ] `/src/app/api/github/repos/[owner]/[repo]/pulls/[number]/merge/route.ts`
- [ ] `/src/app/api/epics/tasks/route.ts`
- [ ] `/src/app/api/epics/tasks/[id]/route.ts`

**Files to Update (4)**:
- [ ] `/src/services/WorkLogManager.ts` - Update API_BASE constant
- [ ] `/src/services/githubApi.ts` - Add backend merge proxy method
- [ ] `next.config.ts` - Add static export config
- [ ] `.env.production` - Create with production URLs

**Files to Create (1)**:
- [ ] `.env.production` - Production environment variables

### Backend Endpoints to Implement

**Work Logs (2 endpoints)**:
- [ ] `GET /api/v1/work-logs?repository={repo}&date={date}` - Retrieve work logs
- [ ] `POST /api/v1/work-logs/entry` - Create work log entry

**GitHub Integration (1 endpoint)**:
- [ ] `PUT /api/v1/github/repos/{owner}/{repo}/pulls/{number}/merge` - Merge PR

**Epic Tasks (2 endpoints)**:
- [ ] `GET/POST /api/v1/epics/tasks` - List/create epic tasks
- [ ] `GET/PUT/DELETE /api/v1/epics/tasks/{id}` - Epic task CRUD

### CORS Configuration Example

```go
// Backend CORS middleware (Go/Gin)
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "http://my-bucket.s3-website-ap-south-1.amazonaws.com")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### S3 Bucket Policy Example

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::my-workflow-app/*"
    }
  ]
}
```

### Testing Checklist

**CORS Testing**:
- [ ] Preflight OPTIONS request succeeds
- [ ] GET requests return correct CORS headers
- [ ] POST requests with JSON body succeed
- [ ] PUT/DELETE requests succeed
- [ ] Credentials included in requests

**Feature Testing**:
- [ ] Task creation works
- [ ] Task status polling works
- [ ] Work log creation works
- [ ] Work log retrieval works
- [ ] Epic task creation works
- [ ] Epic task update works
- [ ] GitHub PR merge works
- [ ] Authentication persists across reloads

**Performance Testing**:
- [ ] FCP ≤1.5s on 3G
- [ ] TTI ≤3s on 3G
- [ ] API response time ≤500ms
- [ ] Static asset caching works
- [ ] No unnecessary re-renders

---

**Document Status**: Draft
**Last Updated**: 2025-10-02T00:19:41Z
**Next Steps**: Review PRD → Run `/pm:prd-parse s3-static-hosting` to create implementation epic
