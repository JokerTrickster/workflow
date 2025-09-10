---
issue: 28
title: Repository Connection State Management
epic: github-repo-integration
created: 2025-09-08T09:28:21Z
priority: critical
effort: 2 days
dependencies: none
---

# Issue #28 Analysis: Repository Connection State Management

## Task Summary
Update existing repository connection flow to persist connected status and enable repository selection for GitHub integration.

## Work Stream Analysis

### Stream A: Component Updates (Frontend)
**Scope**: UI component modifications
**Agent**: frontend-architect
**Files**: 
- `frontend/src/components/RepositoryCard.tsx`
- `frontend/src/components/SearchFilter.tsx`

**Work**:
- Add connection toggle to RepositoryCard
- Add connection status filter to SearchFilter 
- Implement visual states (connected/disconnected)

**Dependencies**: None - can start immediately

### Stream B: State Management (Frontend)
**Scope**: Data layer and state management
**Agent**: frontend-architect  
**Files**:
- `frontend/src/hooks/useRepositories.ts`
- `frontend/src/types/database.ts` (if needed)

**Work**:
- Update useRepositories hook for `is_connected` state
- Implement React Query mutations for connection updates
- Persist connection status in repository data model

**Dependencies**: None - can start immediately

### Stream C: Integration Layer (Backend)
**Scope**: Backend API support (if needed)
**Agent**: backend-architect
**Files**:
- `backend/internal/handlers/repository.go` (potential)
- Database schema validation

**Work**:
- Verify existing Repository model supports `is_connected` field
- Add/update API endpoints for connection status if needed
- Ensure proper data persistence

**Dependencies**: Stream B analysis - may not be needed if frontend-only

## Parallel Execution Plan

### Phase 1: Immediate Start (Parallel)
1. **Stream A**: Component visual updates (frontend-architect)
2. **Stream B**: State management implementation (frontend-architect)

### Phase 2: Validation & Integration  
3. **Stream C**: Backend verification (backend-architect) - only if needed
4. Integration testing of connection flow

## Risk Assessment

**Low Risk**:
- Well-defined scope with existing patterns
- Leverages existing Repository entity
- No breaking changes expected

**Mitigation**:
- Use existing React Query patterns
- Follow established component conventions
- Gradual rollout with feature flags if needed

## Success Criteria

- [ ] Connection toggle visible in RepositoryCard
- [ ] Toggle persists `is_connected` state
- [ ] SearchFilter shows connection status options  
- [ ] Connected repositories trigger 3-tab interface
- [ ] State survives page refresh
- [ ] No breaking changes to existing functionality

## Files to Monitor
```
frontend/src/components/RepositoryCard.tsx
frontend/src/components/SearchFilter.tsx  
frontend/src/hooks/useRepositories.ts
frontend/src/types/database.ts
```

## Coordination Notes
- Both streams can work in parallel
- Stream A focuses on UI/UX
- Stream B focuses on data/state
- Integration point: RepositoryCard using updated hook