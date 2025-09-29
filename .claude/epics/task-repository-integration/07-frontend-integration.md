---
name: frontend-integration
epic: task-repository-integration
type: ui
priority: medium
status: backlog
created: 2025-09-29T02:57:03Z
estimated_days: 2
complexity: low
dependencies: [database-schema-update, task-pipeline-enhancement]
---

# Task: Frontend Integration

## Overview
Update the existing frontend task management interface to display GitHub integration features including issue links, PR URLs, branch information, and Git workflow status. This provides users with full visibility into the automated Git workflow without requiring new components.

## Acceptance Criteria

### Enhanced Task Display
- [ ] **GitHub Issue Links**: Display clickable links to GitHub issues in task history
- [ ] **PR Integration**: Show pull request URLs and status in completed tasks
- [ ] **Branch Information**: Display the feature branch name used for task execution
- [ ] **Git Workflow Status**: Show current status of Git operations (creating issue, executing, creating PR, cleanup)

### Status Indicators
- [ ] **Visual Status Icons**: Clear icons for different Git workflow stages
- [ ] **Progress Tracking**: Real-time updates during Git workflow execution
- [ ] **Error States**: Visual indication when Git operations fail with fallback information
- [ ] **Completion States**: Clear indication of successful Git workflow completion

### User Experience
- [ ] **Minimal UI Changes**: Leverage existing `ClaudeTaskRunner` component structure
- [ ] **Progressive Enhancement**: Git features enhance existing functionality without breaking changes
- [ ] **Responsive Design**: Git workflow information displays properly on all screen sizes
- [ ] **Accessibility**: All new elements meet accessibility standards

### Data Integration
- [ ] **API Updates**: Extend existing API responses to include Git workflow data
- [ ] **Real-time Updates**: WebSocket or polling integration for live status updates
- [ ] **Error Handling**: Graceful handling when Git workflow data is unavailable
- [ ] **Backward Compatibility**: Interface works correctly when Git features are disabled

## Implementation Details

### Enhanced Task History Component
```tsx
// Update existing ClaudeTaskRunner component
interface TaskHistoryItem {
  // Existing fields
  request_id: string;
  status: string;
  tasks: string;
  repository_name: string;
  created_at: string;
  completed_at?: string;
  processing_time_ms?: number;
  result?: string;
  error?: string;

  // New Git workflow fields
  github_issue_url?: string;
  github_pr_url?: string;
  branch_name?: string;
  cleanup_status?: 'pending' | 'in_progress' | 'completed' | 'failed';
  git_operations?: GitOperation[];
}

interface GitOperation {
  operation: string;
  status: 'success' | 'failed' | 'in_progress';
  timestamp: string;
  error?: string;
}

const GitWorkflowStatus: React.FC<{ task: TaskHistoryItem }> = ({ task }) => {
  return (
    <div className="git-workflow-status">
      {task.github_issue_url && (
        <div className="git-workflow-item">
          <GitHubIcon className="w-4 h-4" />
          <a
            href={task.github_issue_url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 hover:text-blue-800 underline"
          >
            GitHub Issue
          </a>
        </div>
      )}

      {task.branch_name && (
        <div className="git-workflow-item">
          <BranchIcon className="w-4 h-4" />
          <span className="text-gray-600 font-mono text-sm">
            {task.branch_name}
          </span>
        </div>
      )}

      {task.github_pr_url && (
        <div className="git-workflow-item">
          <PullRequestIcon className="w-4 h-4" />
          <a
            href={task.github_pr_url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-green-600 hover:text-green-800 underline"
          >
            Pull Request
          </a>
        </div>
      )}

      <CleanupStatus status={task.cleanup_status} />
    </div>
  );
};
```

### Status Indicators Component
```tsx
const GitWorkflowProgress: React.FC<{ task: TaskHistoryItem }> = ({ task }) => {
  const getWorkflowStage = (task: TaskHistoryItem): WorkflowStage => {
    if (task.status === 'pending') return 'queued';
    if (task.status === 'processing') {
      if (!task.github_issue_url) return 'creating_issue';
      if (!task.branch_name) return 'creating_branch';
      if (task.status === 'processing') return 'executing_task';
      if (!task.github_pr_url) return 'creating_pr';
    }
    if (task.status === 'completed') {
      if (task.cleanup_status === 'pending') return 'cleanup_pending';
      if (task.cleanup_status === 'in_progress') return 'cleanup_in_progress';
      if (task.cleanup_status === 'completed') return 'completed';
    }
    if (task.status === 'failed') return 'failed';
    return 'unknown';
  };

  const stage = getWorkflowStage(task);

  return (
    <div className="git-workflow-progress">
      <div className="workflow-stages">
        <WorkflowStage
          name="Issue"
          status={getStageStatus(stage, 'creating_issue')}
          icon={<IssueIcon />}
        />
        <WorkflowStage
          name="Branch"
          status={getStageStatus(stage, 'creating_branch')}
          icon={<BranchIcon />}
        />
        <WorkflowStage
          name="Execute"
          status={getStageStatus(stage, 'executing_task')}
          icon={<PlayIcon />}
        />
        <WorkflowStage
          name="PR"
          status={getStageStatus(stage, 'creating_pr')}
          icon={<PullRequestIcon />}
        />
        <WorkflowStage
          name="Cleanup"
          status={getStageStatus(stage, 'cleanup_completed')}
          icon={<CheckIcon />}
        />
      </div>
    </div>
  );
};

const WorkflowStage: React.FC<{
  name: string;
  status: 'pending' | 'active' | 'completed' | 'failed';
  icon: React.ReactNode;
}> = ({ name, status, icon }) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'text-gray-400';
      case 'active': return 'text-blue-600 animate-pulse';
      case 'completed': return 'text-green-600';
      case 'failed': return 'text-red-600';
      default: return 'text-gray-400';
    }
  };

  return (
    <div className={`workflow-stage ${getStatusColor(status)}`}>
      <div className="stage-icon">{icon}</div>
      <div className="stage-name text-xs">{name}</div>
    </div>
  );
};
```

### Enhanced Task Card
```tsx
const TaskHistoryCard: React.FC<{ task: TaskHistoryItem }> = ({ task }) => {
  return (
    <div className="task-card border rounded-lg p-4 mb-4">
      {/* Existing task information */}
      <div className="task-header flex justify-between items-start mb-3">
        <div className="task-info">
          <h3 className="font-semibold">{task.tasks}</h3>
          <p className="text-sm text-gray-600">
            Repository: {task.repository_name}
          </p>
          <p className="text-sm text-gray-500">
            {formatDate(task.created_at)}
          </p>
        </div>
        <StatusBadge status={task.status} />
      </div>

      {/* Git workflow progress */}
      {(task.github_issue_url || task.branch_name || task.github_pr_url) && (
        <div className="git-workflow-section mb-3">
          <h4 className="text-sm font-medium text-gray-700 mb-2">
            Git Workflow
          </h4>
          <GitWorkflowProgress task={task} />
          <GitWorkflowStatus task={task} />
        </div>
      )}

      {/* Existing result/error display */}
      {task.result && (
        <div className="task-result mb-2">
          <pre className="text-sm bg-gray-100 p-2 rounded overflow-x-auto">
            {task.result}
          </pre>
        </div>
      )}

      {task.error && (
        <div className="task-error">
          <div className="text-red-600 text-sm bg-red-50 p-2 rounded">
            {task.error}
          </div>
        </div>
      )}
    </div>
  );
};
```

### API Integration Updates
```typescript
// Update existing API service
export interface TaskHistoryResponse {
  tasks: TaskHistoryItem[];
  total: number;
  page: number;
  limit: number;
}

export const fetchTaskHistory = async (): Promise<TaskHistoryResponse> => {
  const response = await fetch(`${API_BASE_URL}/workflow/history`, {
    headers: {
      'Authorization': `Bearer ${getAuthToken()}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to fetch task history');
  }

  return response.json();
};

// Add real-time updates for Git workflow status
export const subscribeToTaskUpdates = (
  taskId: string,
  onUpdate: (task: TaskHistoryItem) => void
) => {
  // WebSocket implementation for real-time updates
  const ws = new WebSocket(`${WS_BASE_URL}/tasks/${taskId}/updates`);

  ws.onmessage = (event) => {
    const updatedTask = JSON.parse(event.data);
    onUpdate(updatedTask);
  };

  return () => ws.close();
};
```

### Configuration Component
```tsx
const GitWorkflowSettings: React.FC = () => {
  const [settings, setSettings] = useState({
    enableGitWorkflow: true,
    showBranchNames: true,
    autoRefreshStatus: true,
    refreshInterval: 30000,
  });

  return (
    <div className="git-workflow-settings">
      <h3 className="text-lg font-semibold mb-4">Git Workflow Settings</h3>

      <div className="setting-item mb-3">
        <label className="flex items-center">
          <input
            type="checkbox"
            checked={settings.enableGitWorkflow}
            onChange={(e) => setSettings({
              ...settings,
              enableGitWorkflow: e.target.checked
            })}
            className="mr-2"
          />
          Enable Git Workflow Integration
        </label>
      </div>

      <div className="setting-item mb-3">
        <label className="flex items-center">
          <input
            type="checkbox"
            checked={settings.showBranchNames}
            onChange={(e) => setSettings({
              ...settings,
              showBranchNames: e.target.checked
            })}
            className="mr-2"
          />
          Show Branch Names
        </label>
      </div>

      <div className="setting-item mb-3">
        <label className="flex items-center">
          <input
            type="checkbox"
            checked={settings.autoRefreshStatus}
            onChange={(e) => setSettings({
              ...settings,
              autoRefreshStatus: e.target.checked
            })}
            className="mr-2"
          />
          Auto-refresh Status
        </label>
      </div>
    </div>
  );
};
```

### CSS Enhancements
```css
/* Git workflow specific styles */
.git-workflow-status {
  @apply flex flex-wrap gap-2 mt-2;
}

.git-workflow-item {
  @apply flex items-center gap-1 text-sm;
}

.workflow-stages {
  @apply flex items-center justify-between w-full max-w-md;
}

.workflow-stage {
  @apply flex flex-col items-center gap-1;
}

.stage-icon {
  @apply w-6 h-6 rounded-full border-2 flex items-center justify-center;
}

.workflow-stage.text-gray-400 .stage-icon {
  @apply border-gray-300 text-gray-400;
}

.workflow-stage.text-blue-600 .stage-icon {
  @apply border-blue-600 text-blue-600 bg-blue-50;
}

.workflow-stage.text-green-600 .stage-icon {
  @apply border-green-600 text-green-600 bg-green-50;
}

.workflow-stage.text-red-600 .stage-icon {
  @apply border-red-600 text-red-600 bg-red-50;
}

/* Responsive adjustments */
@media (max-width: 640px) {
  .workflow-stages {
    @apply flex-col gap-2;
  }

  .git-workflow-item {
    @apply text-xs;
  }
}
```

## Testing Requirements

### Component Tests
- [ ] Git workflow status display with various data states
- [ ] Progress indicator transitions
- [ ] Link functionality and external navigation
- [ ] Responsive behavior on different screen sizes

### Integration Tests
- [ ] API integration with new Git workflow fields
- [ ] Real-time update functionality
- [ ] Error handling when Git data unavailable
- [ ] Settings persistence and application

### User Acceptance Tests
- [ ] Task history displays Git workflow information correctly
- [ ] External links open properly in new tabs
- [ ] Progress indicators update in real-time
- [ ] Interface remains usable when Git features disabled

## Definition of Done
- [ ] Git workflow information displayed in existing UI
- [ ] External links to GitHub functional
- [ ] Real-time status updates working
- [ ] Responsive design maintained
- [ ] Accessibility requirements met
- [ ] No breaking changes to existing functionality
- [ ] User settings for Git workflow features
- [ ] Error states handled gracefully

## Dependencies
- Database Schema Update (02) - for new data fields
- Task Pipeline Enhancement (03) - for Git workflow data
- Backend API updates to include new fields
- GitHub access for link validation

## Notes
Frontend integration should enhance the existing interface without overwhelming users. Focus on progressive enhancement and ensure the interface degrades gracefully when Git workflow features are unavailable. Keep the design consistent with existing UI patterns.