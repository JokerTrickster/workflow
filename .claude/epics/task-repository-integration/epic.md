---
name: task-repository-integration
status: backlog
created: 2025-09-29T02:55:44Z
progress: 0%
prd: .claude/prds/task-repository-integration.md
github: https://github.com/JokerTrickster/workflow/issues/86
---

# Epic: Task Repository Integration

## Overview
Enhance the existing local-backend task processing system to automatically manage Git workflows, GitHub issues, and pull requests. This implementation leverages the current RabbitMQ-based task execution infrastructure and extends it with GitHub API integration, branch management, and automated PR creation while maintaining backward compatibility.

## Architecture Decisions

**Leverage Existing Infrastructure**
- Extend current `ConcurrentTaskWorker` instead of creating new task processing
- Reuse existing `RepositoryManager` and `claude_cli.go` utilities for Git operations
- Build on existing GitHub OAuth integration in backend for API authentication

**GitHub API Integration Pattern**
- Utilize existing GitHub service structure (`backend/features/github/`)
- Extend current GitHub client for issue and PR management
- Reuse authentication flow from existing GitHub OAuth provider

**Branch Management Strategy**
- Implement within existing `runTasksClaudeUseCase.go` workflow
- Use existing Git utilities but enhance with branch creation/switching
- Maintain current working directory patterns but ensure branch isolation

**Database Integration**
- Extend existing `workflow_histories` table with GitHub issue/PR references
- Add branch tracking and cleanup status fields
- Reuse current audit logging patterns

## Technical Approach

### Frontend Components
- **Minimal Changes Required**: Leverage existing `ClaudeTaskRunner` component
- **GitHub Links Display**: Add issue URL and PR URL to task history display
- **Status Indicators**: Extend current status display with Git workflow status
- **No New Components**: Work within existing task management UI

### Backend Services

**GitHub API Extensions** (Build on existing `/features/github/`)
- Extend `GitHubService.go` with issue and PR management methods
- Reuse existing GitHub token handling and authentication
- Add issue creation, PR creation, and cleanup operations

**Task Processing Enhancements** (Extend existing task pipeline)
- Modify `runTasksClaudeUseCase.go` to include Git workflow steps
- Add pre-execution branch creation and post-execution PR creation
- Integrate with existing task logging and status tracking

**Database Schema Updates**
- Add fields to `workflow_histories`: `github_issue_url`, `github_pr_url`, `branch_name`, `cleanup_status`
- Maintain existing audit trail structure
- No new tables required

### Infrastructure
- **No New Services**: All functionality within existing local-backend
- **Reuse Existing**: RabbitMQ, database, authentication patterns
- **GitHub API Rate Limiting**: Use existing token rotation if available

## Implementation Strategy

**Phase 1: GitHub Issue Integration** (Non-breaking)
- Extend GitHub service with issue creation
- Add issue URL to task tracking
- Optional feature - tasks work normally if GitHub fails

**Phase 2: Branch Management** (Low risk)
- Enhance Git operations in task execution
- Create branches before task execution
- Track branch names in database

**Phase 3: PR Automation** (Full workflow)
- Add PR creation after successful task completion
- Implement automatic cleanup workflows
- Full Git workflow integration

**Rollback Strategy**
- Feature flags for each phase
- Graceful degradation if GitHub API unavailable
- Existing task execution continues unchanged if new features fail

## Task Breakdown Preview

High-level task categories for implementation:

- [ ] **GitHub Service Extension**: Add issue/PR management to existing GitHub service (2-3 days)
- [ ] **Database Schema Update**: Add Git workflow fields to workflow_histories table (1 day)
- [ ] **Task Pipeline Enhancement**: Modify task execution to include Git workflow steps (3-4 days)
- [ ] **Branch Management**: Implement branch creation, switching, and cleanup (2-3 days)
- [ ] **PR Automation**: Add automated PR creation with templates and linking (2-3 days)
- [ ] **Error Handling & Cleanup**: Robust failure handling and resource cleanup (2 days)
- [ ] **Frontend Integration**: Display GitHub links and status in existing UI (1-2 days)
- [ ] **Testing & Documentation**: Comprehensive testing and documentation updates (2-3 days)

## Dependencies

**External Dependencies**
- GitHub API for issue and PR operations
- Existing GitHub OAuth tokens from backend authentication
- Git CLI already available in local-backend environment

**Internal Dependencies**
- Current local-backend task execution system (stable)
- Existing GitHub OAuth integration in backend service
- Database schema migration capability
- RabbitMQ task processing pipeline

**No Team Dependencies**
- Builds entirely on existing infrastructure
- Uses current authentication mechanisms
- Minimal frontend changes required

## Success Criteria (Technical)

**Performance Benchmarks**
- Task execution time increase < 10% due to Git operations
- GitHub API calls complete within 5 seconds or gracefully degrade
- No impact on concurrent task processing capabilities

**Quality Gates**
- 100% backward compatibility with existing task execution
- Zero data loss during Git operations
- Comprehensive error handling for all GitHub API failures

**Acceptance Criteria**
- All tasks automatically create GitHub issues when GitHub API available
- Branch isolation prevents task interference
- Successful tasks automatically create PRs with proper linking
- Failed operations clean up branches and issues appropriately

## Estimated Effort

**Overall Timeline**: 3-4 weeks (15-20 development days)

**Resource Requirements**
- 1 backend developer familiar with existing codebase
- Access to test repositories for Git workflow testing
- GitHub API token with appropriate permissions

**Critical Path Items**
1. GitHub service extension (foundation for all other work)
2. Task pipeline modification (core workflow changes)
3. Branch management implementation (enables isolation)
4. Integration testing (ensures reliability)

**Risk Factors**
- GitHub API rate limiting (mitigated by graceful degradation)
- Git authentication complexity (mitigated by using existing setup)
- Branch conflict handling (mitigated by unique branch naming)

**Simplification Opportunities**
- Reuse all existing infrastructure and patterns
- Minimal database changes (add fields, no new tables)
- No new services or major architectural changes
- Build incrementally on proven task execution pipeline