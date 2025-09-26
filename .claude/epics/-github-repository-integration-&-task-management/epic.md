---
github: "https://github.com/JokerTrickster/workflow/issues/27"
last_sync: "2025-09-26T17:16:05.573138Z"
status: completed

---

## Overview

Implement GitHub repository connection with integrated task management tabs by extending existing GitHub OAuth authentication and leveraging established UI/API patterns. The system already has GitHub OAuth, repository entities, and task management - this epic focuses on connecting repository selection to 3-tab workspace interface and adding missing logout functionality.

## Architecture Decisions

- **Leverage Existing GitHub OAuth**: No additional authentication needed - already implemented with Supabase Auth
- **Extend GitHubApiService**: Add Issues/PR API calls to existing service class
- **Reuse Repository Pattern**: GitHub data already mapped to Repository entity with `is_connected` field
- **Tab Interface**: Utilize existing Radix UI Tabs components for Task/Logs/Dashboard
- **State Management**: Continue with React Query for GitHub API data caching and synchronization
- **Database Schema**: No changes needed - existing Task entity supports repository_id, GitHub integration fields

## Implementation Timeline: 3 weeks

**Week 1 (Foundation):**
- Repository Connection State Management (2 days)
- Logout Functionality Implementation (1 day) 
- GitHub API Service Extension (3 days - start)

**Week 2 (Core Features):**
- GitHub API Service Extension (complete)
- Three-Tab Workspace Interface (3 days)
- GitHub Issues/PR Integration (2 days)

**Week 3 (Enhancement & Polish):**
- Activity Logging System (2 days)
- Repository Dashboard & Statistics (2 days)
- Integration Testing & Error Boundaries (2 days)

## Success Criteria

- Repository connection persists `is_connected` status
- Connected repositories show 3-tab interface (Task/Logs/Dashboard)
- GitHub Issues/PRs display in Task tab with local task creation
- Logs tab shows chronological activity with filtering
- Dashboard displays repository statistics and active work
- Logout button cleanly terminates session and redirects

## Related Files

- PRD: `.claude/prds/github-repo-integration.md`
- Epic: `.claude/epics/github-repo-integration/epic.md`
- Tasks: `.claude/epics/github-repo-integration/tasks.md`

🤖 Generated with [Claude Code](https://claude.ai/code)