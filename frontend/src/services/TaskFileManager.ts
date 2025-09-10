import { Task } from '../domain/entities/Task';

export interface TaskFileMetadata {
  id: string;
  title: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  repository: string;
  epic: string;
  branch?: string;
  createdAt: string;
  updatedAt: string;
  startedAt?: string;
  completedAt?: string;
  tokensUsed: number;
  githubIssue?: number;
  prUrl?: string;
  buildStatus?: 'pending' | 'success' | 'failure';
  lintStatus?: 'pending' | 'success' | 'failure';
}

export interface TaskFile {
  metadata: TaskFileMetadata;
  content: string;
}

export class TaskFileManager {
  private static instance: TaskFileManager;
  private taskCache: Map<string, TaskFile> = new Map();
  private lastCacheUpdate = 0;
  private readonly CACHE_TTL = 5000; // 5 seconds

  private constructor() {}

  static getInstance(): TaskFileManager {
    if (!TaskFileManager.instance) {
      TaskFileManager.instance = new TaskFileManager();
    }
    return TaskFileManager.instance;
  }

  async loadTasksFromEpics(): Promise<Task[]> {
    try {
      // Check cache validity
      const now = Date.now();
      if (now - this.lastCacheUpdate < this.CACHE_TTL && this.taskCache.size > 0) {
        return this.convertTaskFilesToTasks(Array.from(this.taskCache.values()));
      }

      // Load tasks from backend API
      const response = await fetch('/api/epics/tasks');
      if (!response.ok) {
        throw new Error(`Failed to load tasks: ${response.statusText}`);
      }

      const taskFiles: TaskFile[] = await response.json();
      
      // Update cache
      this.taskCache.clear();
      taskFiles.forEach(taskFile => {
        this.taskCache.set(taskFile.metadata.id, taskFile);
      });
      this.lastCacheUpdate = now;

      return this.convertTaskFilesToTasks(taskFiles);
    } catch (error) {
      console.error('Failed to load tasks from epics:', error);
      return [];
    }
  }

  async createTaskFile(taskData: Omit<TaskFileMetadata, 'id' | 'createdAt' | 'updatedAt'>, content: string): Promise<TaskFile> {
    const taskFile: TaskFile = {
      metadata: {
        id: this.generateTaskId(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        ...taskData,
      },
      content,
    };

    try {
      const response = await fetch('/api/epics/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(taskFile),
      });

      if (!response.ok) {
        throw new Error(`Failed to create task file: ${response.statusText}`);
      }

      const createdTaskFile: TaskFile = await response.json();
      
      // Update cache
      this.taskCache.set(createdTaskFile.metadata.id, createdTaskFile);
      
      return createdTaskFile;
    } catch (error) {
      console.error('Failed to create task file:', error);
      throw error;
    }
  }

  async updateTaskFile(taskId: string, updates: Partial<TaskFileMetadata>, content?: string): Promise<TaskFile> {
    try {
      const existingTask = this.taskCache.get(taskId);
      if (!existingTask) {
        throw new Error(`Task ${taskId} not found in cache`);
      }

      const updatedTaskFile: TaskFile = {
        metadata: {
          ...existingTask.metadata,
          ...updates,
          updatedAt: new Date().toISOString(),
        },
        content: content || existingTask.content,
      };

      const response = await fetch(`/api/epics/tasks/${taskId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedTaskFile),
      });

      if (!response.ok) {
        throw new Error(`Failed to update task file: ${response.statusText}`);
      }

      const result: TaskFile = await response.json();
      
      // Update cache
      this.taskCache.set(taskId, result);
      
      return result;
    } catch (error) {
      console.error('Failed to update task file:', error);
      throw error;
    }
  }

  async getTaskFile(taskId: string): Promise<TaskFile | null> {
    try {
      // Check cache first
      const cached = this.taskCache.get(taskId);
      if (cached) {
        return cached;
      }

      // Fetch from API
      const response = await fetch(`/api/epics/tasks/${taskId}`);
      if (!response.ok) {
        if (response.status === 404) {
          return null;
        }
        throw new Error(`Failed to get task file: ${response.statusText}`);
      }

      const taskFile: TaskFile = await response.json();
      
      // Update cache
      this.taskCache.set(taskId, taskFile);
      
      return taskFile;
    } catch (error) {
      console.error('Failed to get task file:', error);
      return null;
    }
  }

  private convertTaskFilesToTasks(taskFiles: TaskFile[]): Task[] {
    return taskFiles.map(taskFile => this.convertTaskFileToTask(taskFile));
  }

  private convertTaskFileToTask(taskFile: TaskFile): Task {
    const { metadata } = taskFile;
    
    return {
      id: metadata.id,
      title: metadata.title,
      description: taskFile.content.split('\n').slice(0, 3).join(' ').substring(0, 200) + '...',
      status: metadata.status,
      repository_id: 1, // TODO: Get from context
      created_at: metadata.createdAt,
      updated_at: metadata.updatedAt,
      started_at: metadata.startedAt,
      completed_at: metadata.completedAt,
      branch_name: metadata.branch,
      pr_url: metadata.prUrl,
      build_status: metadata.buildStatus,
      lint_status: metadata.lintStatus,
      ai_tokens_used: metadata.tokensUsed,
    };
  }

  private generateTaskId(): string {
    return `task-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
  }

  // Utility method to clear cache
  clearCache(): void {
    this.taskCache.clear();
    this.lastCacheUpdate = 0;
  }
}