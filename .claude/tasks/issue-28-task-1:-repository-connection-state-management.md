---
github: "https://github.com/JokerTrickster/workflow/issues/28"
last_sync: "2025-09-26T17:16:15.264740Z"
status: completed

---

**Epic**: #27 GitHub Repository Integration & Task Management
**Priority**: Critical  
**Effort**: 2 days  
**Dependencies**: None  

## Description
Update existing repository connection flow to persist connected status and enable repository selection for GitHub integration.

## Acceptance Criteria
- [ ] Extend `RepositoryCard` component to show connection toggle
- [ ] Update `useRepositories` hook to manage `is_connected` state  
- [ ] Persist connection status in existing repository data model
- [ ] Add connection status filter to `SearchFilter` component
- [ ] Connected repositories trigger 3-tab workspace interface

## Technical Details
- **Files**: `components/RepositoryCard.tsx`, `hooks/useRepositories.ts`
- **Pattern**: Extend existing React Query mutation for repository updates
- **State**: Update Repository entity `is_connected` boolean field

🤖 Generated with [Claude Code](https://claude.ai/code)