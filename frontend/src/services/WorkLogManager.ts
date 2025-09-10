/**
 * Work Log Manager Service
 * Creates and manages automated work logs in .claude/logs/ directory
 * Designed for git-based synchronization across machines
 */

export interface WorkLogEntry {
  timestamp: string;
  taskId: string;
  taskTitle: string;
  repository: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  progressUpdate?: string;
  issuesDiscovered?: string[];
  improvementsMade?: string[];
  metadata?: {
    branch?: string;
    githubIssue?: number;
    prUrl?: string;
    tokensUsed?: number;
  };
}

export interface DailyWorkLog {
  date: string;
  repository: string;
  entries: WorkLogEntry[];
}

export class WorkLogManager {
  private static instance: WorkLogManager;
  private readonly API_BASE = '/api/work-logs';

  private constructor() {}

  static getInstance(): WorkLogManager {
    if (!WorkLogManager.instance) {
      WorkLogManager.instance = new WorkLogManager();
    }
    return WorkLogManager.instance;
  }

  /**
   * Log task creation
   */
  async logTaskCreated(
    taskId: string,
    taskTitle: string,
    repository: string,
    metadata?: {
      branch?: string;
      githubIssue?: number;
      description?: string;
    }
  ): Promise<void> {
    const entry: WorkLogEntry = {
      timestamp: new Date().toISOString(),
      taskId,
      taskTitle,
      repository,
      status: 'pending',
      progressUpdate: 'Task created automatically',
      metadata: {
        branch: metadata?.branch,
        githubIssue: metadata?.githubIssue,
      }
    };

    await this.writeLogEntry(repository, entry);
  }

  /**
   * Log task status change
   */
  async logTaskStatusChange(
    taskId: string,
    taskTitle: string,
    repository: string,
    status: WorkLogEntry['status'],
    progressUpdate?: string,
    issuesDiscovered?: string[],
    improvementsMade?: string[]
  ): Promise<void> {
    const entry: WorkLogEntry = {
      timestamp: new Date().toISOString(),
      taskId,
      taskTitle,
      repository,
      status,
      progressUpdate,
      issuesDiscovered,
      improvementsMade,
    };

    await this.writeLogEntry(repository, entry);
  }

  /**
   * Log progress update during task execution
   */
  async logProgress(
    taskId: string,
    taskTitle: string,
    repository: string,
    progressUpdate: string,
    metadata?: {
      tokensUsed?: number;
      prUrl?: string;
    }
  ): Promise<void> {
    const entry: WorkLogEntry = {
      timestamp: new Date().toISOString(),
      taskId,
      taskTitle,
      repository,
      status: 'in_progress',
      progressUpdate,
      metadata
    };

    await this.writeLogEntry(repository, entry);
  }

  /**
   * Log issues discovered during task execution
   */
  async logIssuesDiscovered(
    taskId: string,
    taskTitle: string,
    repository: string,
    issues: string[]
  ): Promise<void> {
    const entry: WorkLogEntry = {
      timestamp: new Date().toISOString(),
      taskId,
      taskTitle,
      repository,
      status: 'in_progress',
      progressUpdate: `Issues discovered: ${issues.length} items`,
      issuesDiscovered: issues,
    };

    await this.writeLogEntry(repository, entry);
  }

  /**
   * Log improvements made during task execution
   */
  async logImprovements(
    taskId: string,
    taskTitle: string,
    repository: string,
    improvements: string[]
  ): Promise<void> {
    const entry: WorkLogEntry = {
      timestamp: new Date().toISOString(),
      taskId,
      taskTitle,
      repository,
      status: 'in_progress',
      progressUpdate: `Improvements implemented: ${improvements.length} items`,
      improvementsMade: improvements,
    };

    await this.writeLogEntry(repository, entry);
  }

  /**
   * Write log entry to file system via API
   */
  private async writeLogEntry(repository: string, entry: WorkLogEntry): Promise<void> {
    try {
      const response = await fetch(`${this.API_BASE}/entry`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          repository,
          entry,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to write log entry: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Failed to write work log entry:', error);
      // Don't throw - logging should not break the main workflow
    }
  }

  /**
   * Get work logs for a repository and date range
   */
  async getWorkLogs(
    repository: string,
    startDate?: string,
    endDate?: string
  ): Promise<DailyWorkLog[]> {
    try {
      const queryParams = new URLSearchParams({
        repository,
      });

      if (startDate) queryParams.append('startDate', startDate);
      if (endDate) queryParams.append('endDate', endDate);

      const response = await fetch(`${this.API_BASE}?${queryParams}`);
      
      if (!response.ok) {
        throw new Error(`Failed to get work logs: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Failed to get work logs:', error);
      return [];
    }
  }

  /**
   * Get work logs for a specific task
   */
  async getTaskLogs(
    repository: string,
    taskId: string
  ): Promise<WorkLogEntry[]> {
    try {
      const response = await fetch(`${this.API_BASE}/${taskId}?repository=${encodeURIComponent(repository)}`);
      
      if (!response.ok) {
        throw new Error(`Failed to get task logs: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Failed to get task logs:', error);
      return [];
    }
  }
}