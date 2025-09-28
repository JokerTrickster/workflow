---
github: "https://github.com/JokerTrickster/workflow/issues/30"
last_sync: "2025-09-26T17:16:16.924298Z"
status: completed

---

**Epic**: #27 GitHub Repository Integration & Task Management
**Priority**: High  
**Effort**: 3 days  
**Dependencies**: Task 1 (#28)

## Description
Implement 3-tab workspace interface (Task/Logs/Dashboard) using existing Radix UI Tabs component for connected repositories.

## Acceptance Criteria
- [ ] Extend `WorkspacePanel` to show tabs for connected repositories
- [ ] Create `TaskTab` component with GitHub Issues/PR display
- [ ] Create `LogsTab` component with activity history
- [ ] Create `DashboardTab` component with repository statistics
- [ ] Tab state persistence across navigation
- [ ] Responsive design for mobile/desktop

## Technical Details
- **Files**: `components/WorkspacePanel.tsx`, `components/tabs/TaskTab.tsx`, `components/tabs/LogsTab.tsx`, `components/tabs/DashboardTab.tsx`
- **Pattern**: Use existing Radix UI `Tabs`, `TabsList`, `TabsTrigger`, `TabsContent`
- **State**: Local state for active tab, React Query for tab data

🤖 Generated with [Claude Code](https://claude.ai/code)