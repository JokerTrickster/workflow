---
issue: 30
title: Three-Tab Workspace Interface
epic: github-repo-integration
created: 2025-09-08T09:54:41Z
priority: high
effort: 3 days
dependencies: Task 1 (#28)
---

# Issue #30 Analysis: Three-Tab Workspace Interface

## Task Summary
Implement 3-tab workspace interface (Task/Logs/Dashboard) using existing Radix UI Tabs component for connected repositories.

## Work Stream Analysis

### Stream A: Core Tab Infrastructure (Frontend)
**Scope**: Main tab interface and workspace panel extension
**Agent**: frontend-architect
**Files**: 
- `frontend/src/presentation/components/WorkspacePanel.tsx`
- Tab infrastructure setup

**Work**:
- Extend `WorkspacePanel` to show tabs for connected repositories
- Implement Radix UI Tabs infrastructure
- Tab state persistence across navigation
- Responsive design for mobile/desktop
- Integration with repository connection state

**Dependencies**: Task #28 completed (repository connection state) ✅

### Stream B: Task Tab Component (Frontend)
**Scope**: GitHub Issues and PR display tab
**Agent**: frontend-architect
**Files**:
- `frontend/src/presentation/components/tabs/TaskTab.tsx`
- Related UI components for Issue/PR display

**Work**:
- Create `TaskTab` component with GitHub Issues/PR display
- Integrate with GitHub API service (from Task #29) ✅
- Issue and PR listing with proper UI
- Filter and search functionality
- Task creation integration

**Dependencies**: Task #29 completed (GitHub API service) ✅

### Stream C: Support Tab Components (Frontend)
**Scope**: Logs and Dashboard tab implementation
**Agent**: frontend-architect
**Files**:
- `frontend/src/presentation/components/tabs/LogsTab.tsx`
- `frontend/src/presentation/components/tabs/DashboardTab.tsx`

**Work**:
- Create `LogsTab` component with activity history
- Create `DashboardTab` component with repository statistics
- Activity logging integration
- Statistics and progress visualization

**Dependencies**: Stream A completion (tab infrastructure)

## Parallel Execution Plan

### Phase 1: Foundation 
1. **Stream A**: Core tab infrastructure and WorkspacePanel extension

### Phase 2: Tab Implementation (Parallel)
2. **Stream B**: TaskTab component with GitHub integration
3. **Stream C**: LogsTab and DashboardTab components

## Technical Architecture Analysis

### Current WorkspacePanel Structure
Based on codebase analysis:
- Uses Radix UI components throughout
- Has existing workspace interface patterns
- Connected to repository state management
- Follows established component conventions

### Integration Points
- **Repository State**: From Task #28 - `is_connected` status and connection management
- **GitHub API**: From Task #29 - Issues and PR data fetching
- **Tab Infrastructure**: Radix UI `Tabs`, `TabsList`, `TabsTrigger`, `TabsContent`
- **State Management**: React Query for data, local state for UI

### Radix UI Tabs Pattern
```tsx
<Tabs defaultValue="tasks">
  <TabsList>
    <TabsTrigger value="tasks">Tasks</TabsTrigger>
    <TabsTrigger value="logs">Logs</TabsTrigger>
    <TabsTrigger value="dashboard">Dashboard</TabsTrigger>
  </TabsList>
  <TabsContent value="tasks">
    <TaskTab />
  </TabsContent>
  <TabsContent value="logs">
    <LogsTab />
  </TabsContent>
  <TabsContent value="dashboard">
    <DashboardTab />
  </TabsContent>
</Tabs>
```

## Risk Assessment

**Medium Risk**:
- Complex tab state management
- Integration with multiple previous tasks
- Responsive design across different screen sizes

**Mitigation**:
- Use established Radix UI patterns from existing codebase
- Leverage completed infrastructure from Tasks #28 and #29
- Progressive enhancement approach
- Mobile-first responsive design

## Success Criteria

- [ ] `WorkspacePanel` shows 3-tab interface for connected repositories
- [ ] `TaskTab` displays GitHub Issues and PRs with proper UI
- [ ] `LogsTab` shows activity history with filtering
- [ ] `DashboardTab` displays repository statistics and progress
- [ ] Tab state persists across navigation (localStorage or URL)
- [ ] Responsive design works on mobile and desktop
- [ ] Only connected repositories show the tab interface
- [ ] All tabs use existing design system and patterns
- [ ] Integration with GitHub API data from Task #29
- [ ] Connection state management from Task #28

## Files to Monitor
```
frontend/src/presentation/components/WorkspacePanel.tsx
frontend/src/presentation/components/tabs/TaskTab.tsx
frontend/src/presentation/components/tabs/LogsTab.tsx
frontend/src/presentation/components/tabs/DashboardTab.tsx
```

## Coordination Notes
- Stream A creates the foundation that Streams B and C build upon
- Stream B and C can work in parallel once Stream A is complete
- All streams should follow existing Radix UI and component conventions
- Integration points: repository connection state, GitHub API data
- Tab state should be consistent and persistent for good UX