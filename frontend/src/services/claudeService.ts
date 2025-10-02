import { apiClient } from '../infrastructure/api/ApiClient';

// Request interface matching the backend structure
export interface ReqRunTasksClaude {
  tasks: string;           // 실행할 작업 내용
  repository_name: string; // 레포지토리 이름 (필수)
  provider: 'claude' | 'codex' | 'cursor'; // AI 제공자 (필수)
  working_dir?: string;    // 작업 디렉토리 (옵션)
  interactive?: boolean;   // 대화형 모드: 여러 작업을 순차 실행
  cmd?: string;           // 명령어 (옵션)
  continue_task?: boolean; // 기존 작업 이어서 하기 (옵션)
}

// Response interface from the backend
export interface ClaudeTaskResponse {
  request_id: string;
  status: string;
  message: string;
  created_at: string;
}

// Task status interface for polling
export interface TaskStatus {
  request_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled' | 'timeout';
  result?: any;
  error?: string;
  processing_time_ms?: number;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
}

// Task history interfaces matching backend models
export interface WorkflowHistory {
  id: number;
  request_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled';
  tasks: string;
  repository_name: string;
  provider: 'claude' | 'codex' | 'cursor';
  working_dir?: string;
  cmd?: string;
  interactive: boolean;
  continue_task?: boolean;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  processing_time_ms?: number;
  result?: string;
  error?: string;

  // GitHub Integration Fields
  github_issue_url?: string;
  github_pr_url?: string;
  branch_name?: string;
  cleanup_status?: string;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface TaskHistoryResponse {
  tasks: WorkflowHistory[];
  total_count: number;
  page: number;
  limit: number;
  has_more: boolean;
}

export interface TaskHistoryParams {
  page?: number;
  limit?: number;
}

/**
 * Claude service for handling task execution requests
 */
export class ClaudeService {
  private static instance: ClaudeService;

  public static getInstance(): ClaudeService {
    if (!ClaudeService.instance) {
      ClaudeService.instance = new ClaudeService();
    }
    return ClaudeService.instance;
  }

  /**
   * Submit a task to Claude for execution
   */
  async runTasks(request: ReqRunTasksClaude): Promise<ClaudeTaskResponse> {
    try {
      const response = await apiClient.post<ClaudeTaskResponse>('/tasks', request as unknown as Record<string, unknown>);
      return response;
    } catch (error) {
      console.error('Failed to submit Claude task:', error);
      throw error;
    }
  }

  /**
   * Get status of a running task
   */
  async getTaskStatus(requestId: string): Promise<TaskStatus> {
    try {
      const response = await apiClient.get<TaskStatus>(`/tasks/${requestId}/status`);
      return response;
    } catch (error) {
      console.error('Failed to get task status:', error);
      throw error;
    }
  }

  /**
   * Cancel a running task
   */
  async cancelTask(requestId: string): Promise<void> {
    try {
      await apiClient.delete(`/tasks/${requestId}`);
    } catch (error) {
      console.error('Failed to cancel task:', error);
      throw error;
    }
  }

  /**
   * Poll task status until completion
   */
  async pollTaskStatus(
    requestId: string,
    onProgress?: (status: TaskStatus) => void,
    pollInterval: number = 2000
  ): Promise<TaskStatus> {
    return new Promise((resolve, reject) => {
      const poll = async () => {
        try {
          const status = await this.getTaskStatus(requestId);
          
          if (onProgress) {
            onProgress(status);
          }

          // Check if task is completed
          if (['completed', 'failed', 'cancelled', 'timeout'].includes(status.status)) {
            resolve(status);
            return;
          }

          // Continue polling
          setTimeout(poll, pollInterval);
        } catch (error) {
          reject(error);
        }
      };

      poll();
    });
  }


  /**
   * Get task history for a repository with pagination
   */
  async getTaskHistory(
    repositoryName: string,
    params: TaskHistoryParams = {}
  ): Promise<TaskHistoryResponse> {
    try {
      const queryParams = new URLSearchParams();
      queryParams.append('repository_name', repositoryName);
      if (params.page) queryParams.append('page', params.page.toString());
      if (params.limit) queryParams.append('limit', params.limit.toString());

      const endpoint = `/tasks?${queryParams.toString()}`;

      const response = await apiClient.get<TaskHistoryResponse>(endpoint);
      return response;
    } catch (error) {
      console.error('Failed to get task history:', error);
      throw error;
    }
  }
}

// Export singleton instance
export const claudeService = ClaudeService.getInstance();