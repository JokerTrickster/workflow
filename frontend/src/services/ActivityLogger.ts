/**
 * Activity Logger Service
 * Manages activity logging with localStorage persistence and filtering capabilities
 */

import { 
  ActivityLog, 
  ActivityType, 
  ActivityLevel, 
  ActivityLogMetadata, 
  ActivityLogFilter,
  ActivityLogExportOptions,
  ACTIVITY_EVENTS,
  ActivityEventType
} from '../types/activity';

export class ActivityLogger {
  private static readonly STORAGE_KEY = 'workflow_activity_logs';
  private static readonly MAX_LOGS = 1000; // Limit logs to prevent localStorage overflow
  private static instance: ActivityLogger | null = null;
  
  private logs: ActivityLog[] = [];
  private listeners: ((logs: ActivityLog[]) => void)[] = [];

  private constructor() {
    this.loadLogs();
    this.cleanupOldLogs();
  }

  public static getInstance(): ActivityLogger {
    if (!ActivityLogger.instance) {
      ActivityLogger.instance = new ActivityLogger();
    }
    return ActivityLogger.instance;
  }

  /**
   * Subscribe to log updates
   */
  public subscribe(listener: (logs: ActivityLog[]) => void): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  /**
   * Notify all listeners of log updates
   */
  private notifyListeners(): void {
    this.listeners.forEach(listener => listener([...this.logs]));
  }

  /**
   * Load logs from localStorage
   */
  private loadLogs(): void {
    if (typeof window === 'undefined') {
      this.logs = [];
      return;
    }
    
    try {
      const storedLogs = localStorage.getItem(ActivityLogger.STORAGE_KEY);
      if (storedLogs) {
        this.logs = JSON.parse(storedLogs);
        // Ensure logs are sorted by timestamp (newest first)
        this.logs.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
      }
    } catch (error) {
      console.error('Failed to load activity logs from localStorage:', error);
      this.logs = [];
    }
  }

  /**
   * Save logs to localStorage
   */
  private saveLogs(): void {
    if (typeof window === 'undefined') {
      return;
    }
    
    try {
      localStorage.setItem(ActivityLogger.STORAGE_KEY, JSON.stringify(this.logs));
    } catch (error) {
      console.error('Failed to save activity logs to localStorage:', error);
      // If storage is full, try to clean up and save again
      if (error instanceof Error && error.name === 'QuotaExceededError') {
        this.cleanupOldLogs(true);
        try {
          localStorage.setItem(ActivityLogger.STORAGE_KEY, JSON.stringify(this.logs));
        } catch (retryError) {
          console.error('Failed to save activity logs after cleanup:', retryError);
        }
      }
    }
  }

  /**
   * Clean up old logs to maintain storage limit
   */
  private cleanupOldLogs(aggressive = false): void {
    const maxLogs = aggressive ? Math.floor(ActivityLogger.MAX_LOGS * 0.5) : ActivityLogger.MAX_LOGS;
    
    if (this.logs.length > maxLogs) {
      // Keep most recent logs
      this.logs = this.logs
        .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
        .slice(0, maxLogs);
      
      this.saveLogs();
    }

    // Also remove logs older than 30 days
    const thirtyDaysAgo = new Date();
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
    
    const initialLength = this.logs.length;
    this.logs = this.logs.filter(log => new Date(log.timestamp) > thirtyDaysAgo);
    
    if (this.logs.length !== initialLength) {
      this.saveLogs();
    }
  }

  /**
   * Generate unique ID for log entries
   */
  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
  }

  /**
   * Log a generic activity
   */
  public log(
    type: ActivityType,
    level: ActivityLevel,
    title: string,
    description: string,
    metadata?: ActivityLogMetadata
  ): void {
    const logEntry: ActivityLog = {
      id: this.generateId(),
      timestamp: new Date().toISOString(),
      type,
      level,
      title,
      description,
      metadata
    };

    // Add to beginning of array (newest first)
    this.logs.unshift(logEntry);
    
    // Clean up if needed
    if (this.logs.length > ActivityLogger.MAX_LOGS) {
      this.logs = this.logs.slice(0, ActivityLogger.MAX_LOGS);
    }

    this.saveLogs();
    this.notifyListeners();
  }

  /**
   * Repository connection events
   */
  public logRepositoryConnected(repositoryId: number, repositoryName: string, localPath?: string): void {
    this.log(
      'connection',
      'success',
      'Repository connected',
      `Successfully connected to ${repositoryName} repository`,
      {
        repositoryId,
        repositoryName,
        userAction: ACTIVITY_EVENTS.REPOSITORY_CONNECTED,
        context: { localPath }
      }
    );
  }

  public logRepositoryDisconnected(repositoryId: number, repositoryName: string): void {
    this.log(
      'connection',
      'info',
      'Repository disconnected',
      `Disconnected from ${repositoryName} repository`,
      {
        repositoryId,
        repositoryName,
        userAction: ACTIVITY_EVENTS.REPOSITORY_DISCONNECTED
      }
    );
  }

  public logRepositoryConnectionFailed(repositoryName: string, error: string): void {
    this.log(
      'connection',
      'error',
      'Repository connection failed',
      `Failed to connect to ${repositoryName}: ${error}`,
      {
        repositoryName,
        errorMessage: error,
        userAction: ACTIVITY_EVENTS.REPOSITORY_CONNECTION_FAILED
      }
    );
  }

  /**
   * Task-related events
   */
  public logTaskCreated(
    taskId: string,
    taskTitle: string,
    repositoryId: number,
    repositoryName: string,
    metadata?: { githubUrl?: string; branchName?: string }
  ): void {
    this.log(
      'task',
      'info',
      'Task created',
      `New task "${taskTitle}" created for ${repositoryName}`,
      {
        taskId,
        taskTitle,
        repositoryId,
        repositoryName,
        branchName: metadata?.branchName,
        githubUrl: metadata?.githubUrl,
        userAction: ACTIVITY_EVENTS.TASK_CREATED
      }
    );
  }

  public logTaskStarted(taskId: string, taskTitle: string): void {
    this.log(
      'task',
      'info',
      'Task started',
      `Task "${taskTitle}" execution started`,
      {
        taskId,
        taskTitle,
        taskStatus: 'in_progress',
        userAction: ACTIVITY_EVENTS.TASK_STARTED
      }
    );
  }

  public logTaskCompleted(taskId: string, taskTitle: string, duration?: number): void {
    this.log(
      'task',
      'success',
      'Task completed',
      `Task "${taskTitle}" completed successfully`,
      {
        taskId,
        taskTitle,
        taskStatus: 'completed',
        duration,
        userAction: ACTIVITY_EVENTS.TASK_COMPLETED
      }
    );
  }

  public logTaskFailed(taskId: string, taskTitle: string, error: string): void {
    this.log(
      'task',
      'error',
      'Task failed',
      `Task "${taskTitle}" failed: ${error}`,
      {
        taskId,
        taskTitle,
        taskStatus: 'failed',
        errorMessage: error,
        userAction: ACTIVITY_EVENTS.TASK_FAILED
      }
    );
  }

  /**
   * GitHub-related events
   */
  public logGitHubSync(repositoryName: string, type: 'started' | 'completed' | 'failed', details?: { duration?: number; apiCallCount?: number; error?: string }): void {
    const level: ActivityLevel = type === 'failed' ? 'error' : type === 'completed' ? 'success' : 'info';
    const action = type === 'started' ? ACTIVITY_EVENTS.GITHUB_SYNC_STARTED : 
                   type === 'completed' ? ACTIVITY_EVENTS.GITHUB_SYNC_COMPLETED : 
                   ACTIVITY_EVENTS.GITHUB_SYNC_FAILED;

    this.log(
      'github',
      level,
      `GitHub sync ${type}`,
      `GitHub synchronization ${type} for ${repositoryName}`,
      {
        repositoryName,
        userAction: action,
        duration: details?.duration,
        apiCallCount: details?.apiCallCount,
        errorMessage: type === 'failed' ? details?.error : undefined
      }
    );
  }

  public logGitHubApiCall(endpoint: string, method: string, rateLimitRemaining?: number): void {
    this.log(
      'github',
      'info',
      'GitHub API call',
      `${method} ${endpoint}`,
      {
        userAction: ACTIVITY_EVENTS.GITHUB_API_CALL,
        rateLimitRemaining,
        context: { endpoint, method }
      }
    );
  }

  public logGitHubRateLimit(remaining: number, resetTime: string): void {
    const level: ActivityLevel = remaining < 100 ? 'error' : remaining < 500 ? 'warning' : 'info';
    const action = remaining === 0 ? ACTIVITY_EVENTS.GITHUB_RATE_LIMIT_EXCEEDED : ACTIVITY_EVENTS.GITHUB_RATE_LIMIT_WARNING;
    
    this.log(
      'github',
      level,
      'GitHub rate limit',
      `GitHub API rate limit: ${remaining} requests remaining. Resets at ${resetTime}`,
      {
        rateLimitRemaining: remaining,
        userAction: action,
        context: { resetTime }
      }
    );
  }

  /**
   * Navigation events
   */
  public logTabSwitch(previousTab: string, currentTab: string, repositoryName?: string): void {
    this.log(
      'navigation',
      'info',
      'Tab switched',
      repositoryName 
        ? `Switched from ${previousTab} to ${currentTab} tab in ${repositoryName}` 
        : `Switched from ${previousTab} to ${currentTab} tab`,
      {
        previousTab,
        currentTab,
        repositoryName,
        userAction: ACTIVITY_EVENTS.TAB_SWITCHED
      }
    );
  }

  public logWorkspaceAccess(repositoryName: string, action: 'opened' | 'closed'): void {
    this.log(
      'navigation',
      'info',
      `Workspace ${action}`,
      `${action === 'opened' ? 'Accessed' : 'Closed'} workspace for ${repositoryName}`,
      {
        repositoryName,
        userAction: action === 'opened' ? ACTIVITY_EVENTS.WORKSPACE_OPENED : ACTIVITY_EVENTS.WORKSPACE_CLOSED
      }
    );
  }

  /**
   * Get filtered logs
   */
  public getLogs(filter?: ActivityLogFilter): ActivityLog[] {
    let filteredLogs = [...this.logs];

    if (filter) {
      // Filter by type
      if (filter.type && filter.type !== 'all') {
        filteredLogs = filteredLogs.filter(log => log.type === filter.type);
      }

      // Filter by level
      if (filter.level && filter.level !== 'all') {
        filteredLogs = filteredLogs.filter(log => log.level === filter.level);
      }

      // Filter by date range
      if (filter.dateRange) {
        filteredLogs = filteredLogs.filter(log => {
          const logDate = new Date(log.timestamp);
          return logDate >= filter.dateRange!.start && logDate <= filter.dateRange!.end;
        });
      }

      // Filter by search query
      if (filter.searchQuery) {
        const query = filter.searchQuery.toLowerCase();
        filteredLogs = filteredLogs.filter(log => 
          log.title.toLowerCase().includes(query) ||
          log.description.toLowerCase().includes(query) ||
          log.metadata?.repositoryName?.toLowerCase().includes(query) ||
          log.metadata?.taskTitle?.toLowerCase().includes(query)
        );
      }

      // Filter by repository
      if (filter.repositoryId) {
        filteredLogs = filteredLogs.filter(log => 
          log.metadata?.repositoryId === filter.repositoryId
        );
      }
    }

    return filteredLogs;
  }

  /**
   * Export logs
   */
  public exportLogs(options: ActivityLogExportOptions): string {
    const logs = this.getLogs(options.filters);
    
    if (options.format === 'csv') {
      const headers = ['Timestamp', 'Type', 'Level', 'Title', 'Description'];
      if (options.includeMetadata) {
        headers.push('Repository', 'Task', 'Branch', 'Duration', 'Error');
      }

      const csvRows = [
        headers.join(','),
        ...logs.map(log => {
          const row = [
            log.timestamp,
            log.type,
            log.level,
            `"${log.title.replace(/"/g, '""')}"`,
            `"${log.description.replace(/"/g, '""')}"`
          ];

          if (options.includeMetadata) {
            row.push(
              log.metadata?.repositoryName || '',
              log.metadata?.taskTitle || '',
              log.metadata?.branchName || '',
              log.metadata?.duration?.toString() || '',
              log.metadata?.errorMessage?.replace(/"/g, '""') || ''
            );
          }

          return row.join(',');
        })
      ];

      return csvRows.join('\n');
    } else {
      // JSON format
      return JSON.stringify(logs, null, 2);
    }
  }

  /**
   * Clear all logs
   */
  public clearLogs(): void {
    this.logs = [];
    this.saveLogs();
    this.notifyListeners();
  }

  /**
   * Get log statistics
   */
  public getStatistics(): {
    total: number;
    byType: Record<ActivityType, number>;
    byLevel: Record<ActivityLevel, number>;
    recentActivity: number; // Last 24 hours
  } {
    const stats = {
      total: this.logs.length,
      byType: {
        connection: 0,
        task: 0,
        github: 0,
        navigation: 0
      } as Record<ActivityType, number>,
      byLevel: {
        info: 0,
        success: 0,
        warning: 0,
        error: 0
      } as Record<ActivityLevel, number>,
      recentActivity: 0
    };

    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);

    this.logs.forEach(log => {
      stats.byType[log.type]++;
      stats.byLevel[log.level]++;
      
      if (new Date(log.timestamp) > yesterday) {
        stats.recentActivity++;
      }
    });

    return stats;
  }
}