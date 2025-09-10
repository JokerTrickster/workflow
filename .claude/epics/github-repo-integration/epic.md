---
name: github-repo-integration
status: backlog
created: 2025-09-08T09:16:25Z
progress: 0%
prd: .claude/prds/github-repo-integration.md
github: https://github.com/JokerTrickster/workflow/issues/27
---

# Epic: GitHub Repository Integration & Task Management

## Overview

Implement GitHub repository connection with integrated task management tabs by extending existing GitHub OAuth authentication and leveraging established UI/API patterns. The system already has GitHub OAuth, repository entities, and task management - this epic focuses on connecting repository selection to 3-tab workspace interface and adding missing logout functionality.

## Architecture Decisions

- **Leverage Existing GitHub OAuth**: No additional authentication needed - already implemented with Supabase Auth
- **Extend GitHubApiService**: Add Issues/PR API calls to existing service class
- **Reuse Repository Pattern**: GitHub data already mapped to Repository entity with `is_connected` field
- **Tab Interface**: Utilize existing Radix UI Tabs components for Task/Logs/Dashboard
- **State Management**: Continue with React Query for GitHub API data caching and synchronization
- **Database Schema**: No changes needed - existing Task entity supports repository_id, GitHub integration fields

## Technical Approach

### Frontend Components
**Extend Existing Components:**
- `RepositoryCard`: Add repository connection toggle (already has connect/workspace actions)
- `WorkspacePanel`: Implement 3-tab interface using existing Tabs component
- `SearchFilter`: Add connection status filtering (connected/disconnected repositories)

**New Components (Minimal):**
- `TaskTab`: Display GitHub Issues/PRs + local task creation
- `LogsTab`: Activity history with filtering
- `DashboardTab`: Repository statistics and progress overview
- `LogoutButton`: Simple authentication state reset

### Backend Services
**Extend GitHubApiService:**
- Add `fetchRepositoryIssues(repoId)` method
- Add `fetchRepositoryPullRequests(repoId)` method
- Implement rate limiting and error recovery patterns

**Data Models:**
- Repository entity already exists with GitHub integration fields
- Task entity already supports repository linking
- Activity logging through existing workflow audit trail

### Infrastructure
- **GitHub API Rate Limiting**: Implement caching and batch requests
- **Real-time Updates**: Optional WebSocket for live GitHub data sync
- **Error Boundaries**: Graceful degradation when GitHub API unavailable

## Implementation Strategy

### Phase 1: Repository Connection (Week 1)
- Update repository connection flow to persist `is_connected` status
- Add repository selection UI with connection toggle
- Implement connected repository state management

### Phase 2: Tab Interface (Week 2)
- Create 3-tab workspace layout using existing Tabs components
- Implement basic GitHub Issues/PR data fetching and display
- Add local task creation form integration

### Phase 3: Logs & Dashboard (Week 3)
- Implement activity logging for repository and task actions
- Create dashboard with repository statistics and progress metrics
- Add logout functionality to existing auth system

## Task Breakdown (8 Tasks)

Detailed implementation tasks created in: `.claude/epics/github-repo-integration/tasks.md`

**Week 1 (Foundation):**
- [ ] Task 1: Repository Connection State Management (#28) (2 days)
- [ ] Task 7: Logout Functionality Implementation (#34) (1 day) 
- [ ] Task 2: GitHub API Service Extension (#29) (3 days - start)

**Week 2 (Core Features):**
- [ ] Task 2: GitHub API Service Extension (#29) (complete)
- [ ] Task 3: Three-Tab Workspace Interface (#30) (3 days)
- [ ] Task 4: GitHub Issues/PR Integration (#31) (2 days)

**Week 3 (Enhancement & Polish):**
- [ ] Task 5: Activity Logging System (#32) (2 days)
- [ ] Task 6: Repository Dashboard & Statistics (#33) (2 days)
- [ ] Task 8: Integration Testing & Error Boundaries (#35) (2 days)

## Dependencies

### External Dependencies
- **GitHub API v4 (GraphQL)**: Issues and Pull Requests data
- **GitHub OAuth App**: Already configured through Supabase Auth

### Internal Dependencies  
- **Existing GitHubApiService**: Foundation for API extensions
- **Supabase Auth Context**: User session and GitHub token management
- **React Query Setup**: Server state management and caching
- **Radix UI Components**: Tab interface and form components

### Team Dependencies
- **Frontend**: Extend existing components and implement new tab interfaces
- **Backend**: Minimal - primarily frontend-focused with API service extensions
- **DevOps**: No additional infrastructure - uses existing Supabase and GitHub OAuth setup

## Success Criteria (Technical)

### Performance Benchmarks
- GitHub API response time: < 3 seconds (with caching)
- Tab switching response: < 500ms (React Query cache hits)
- Repository connection: < 2 seconds for OAuth flow

### Quality Gates
- Zero breaking changes to existing authentication flow
- All existing repository and task functionality preserved
- Error boundaries prevent GitHub API failures from crashing UI
- Logout completely clears session and redirects to login

### Acceptance Criteria
- Repository connection persists `is_connected` status
- Connected repositories show 3-tab interface (Task/Logs/Dashboard)
- GitHub Issues/PRs display in Task tab with local task creation
- Logs tab shows chronological activity with filtering
- Dashboard displays repository statistics and active work
- Logout button cleanly terminates session and redirects

## Estimated Effort

### Overall Timeline: 3 weeks
- **Week 1**: Repository connection and state management (40% effort)
- **Week 2**: Tab interface and GitHub API integration (40% effort) 
- **Week 3**: Logs, dashboard, logout, and polish (20% effort)

### Resource Requirements
- **1 Frontend Developer**: Primary implementation
- **0.2 Backend Developer**: GitHub API service extensions
- **0.1 DevOps**: OAuth configuration validation

### Critical Path Items
1. Repository connection state management (foundation)
2. GitHub API service extensions (enables all features)
3. Tab interface implementation (user experience core)
4. Integration testing (quality assurance)

**Risk Mitigation**: Heavy reuse of existing patterns minimizes technical risk. Primary risk is GitHub API rate limiting - mitigated through caching and graceful degradation.