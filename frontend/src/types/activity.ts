/**
 * Activity logging system types
 * Defines the structure for activity logs that track user actions and system events
 */

export type ActivityType = 'connection' | 'task' | 'github' | 'navigation';
export type ActivityLevel = 'info' | 'success' | 'warning' | 'error';

export interface ActivityLogMetadata {
  // Task-related metadata
  taskId?: string;
  taskTitle?: string;
  taskStatus?: string;
  
  // Repository-related metadata
  repositoryId?: number;
  repositoryName?: string;
  
  // GitHub-related metadata
  branchName?: string;
  prUrl?: string;
  prNumber?: number;
  issueNumber?: number;
  githubUrl?: string;
  
  // Performance metadata
  duration?: number;
  apiCallCount?: number;
  rateLimitRemaining?: number;
  
  // User metadata
  userId?: string;
  userAction?: string;
  
  // Error metadata
  errorMessage?: string;
  errorCode?: string;
  
  // Navigation metadata
  previousTab?: string;
  currentTab?: string;
  
  // Additional context
  context?: Record<string, unknown>;
}

export interface ActivityLog {
  id: string;
  timestamp: string;
  type: ActivityType;
  level: ActivityLevel;
  title: string;
  description: string;
  metadata?: ActivityLogMetadata;
}

export interface ActivityLogFilter {
  type?: ActivityType | 'all';
  level?: ActivityLevel | 'all';
  dateRange?: {
    start: Date;
    end: Date;
  };
  searchQuery?: string;
  repositoryId?: number;
}

export interface ActivityLogExportOptions {
  format: 'csv' | 'json';
  includeMetadata: boolean;
  filters?: ActivityLogFilter;
}

// Predefined activity event types for consistency
export const ACTIVITY_EVENTS = {
  // Repository connection events
  REPOSITORY_CONNECTED: 'repository_connected',
  REPOSITORY_DISCONNECTED: 'repository_disconnected',
  REPOSITORY_CONNECTION_FAILED: 'repository_connection_failed',
  
  // Task events
  TASK_CREATED: 'task_created',
  TASK_STARTED: 'task_started',
  TASK_COMPLETED: 'task_completed',
  TASK_FAILED: 'task_failed',
  TASK_CANCELLED: 'task_cancelled',
  TASK_UPDATED: 'task_updated',
  
  // GitHub sync events
  GITHUB_SYNC_STARTED: 'github_sync_started',
  GITHUB_SYNC_COMPLETED: 'github_sync_completed',
  GITHUB_SYNC_FAILED: 'github_sync_failed',
  GITHUB_API_CALL: 'github_api_call',
  GITHUB_RATE_LIMIT_WARNING: 'github_rate_limit_warning',
  GITHUB_RATE_LIMIT_EXCEEDED: 'github_rate_limit_exceeded',
  
  // Navigation events
  TAB_SWITCHED: 'tab_switched',
  WORKSPACE_OPENED: 'workspace_opened',
  WORKSPACE_CLOSED: 'workspace_closed',
  
  // System events
  APP_INITIALIZED: 'app_initialized',
  ERROR_OCCURRED: 'error_occurred',
} as const;

export type ActivityEventType = typeof ACTIVITY_EVENTS[keyof typeof ACTIVITY_EVENTS];