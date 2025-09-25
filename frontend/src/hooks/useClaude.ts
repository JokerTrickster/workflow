import { useState, useCallback } from 'react';
import { claudeService, ReqRunTasksClaude, ClaudeTaskResponse, TaskStatus } from '../services/claudeService';

export interface UseClaudeResult {
  // State
  isLoading: boolean;
  isPolling: boolean;
  currentTask: ClaudeTaskResponse | null;
  taskStatus: TaskStatus | null;
  error: string | null;

  // Actions
  runTasks: (request: ReqRunTasksClaude) => Promise<ClaudeTaskResponse>;
  runTasksAndWait: (request: ReqRunTasksClaude) => Promise<TaskStatus>;
  getTaskStatus: (requestId: string) => Promise<TaskStatus>;
  cancelTask: (requestId: string) => Promise<void>;
  clearError: () => void;
  reset: () => void;
}

/**
 * React hook for managing Claude task execution
 */
export function useClaude(): UseClaudeResult {
  const [isLoading, setIsLoading] = useState(false);
  const [isPolling, setIsPolling] = useState(false);
  const [currentTask, setCurrentTask] = useState<ClaudeTaskResponse | null>(null);
  const [taskStatus, setTaskStatus] = useState<TaskStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const reset = useCallback(() => {
    setIsLoading(false);
    setIsPolling(false);
    setCurrentTask(null);
    setTaskStatus(null);
    setError(null);
  }, []);

  const runTasks = useCallback(async (request: ReqRunTasksClaude): Promise<ClaudeTaskResponse> => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await claudeService.runTasks(request);
      setCurrentTask(response);
      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to submit task';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const getTaskStatus = useCallback(async (requestId: string): Promise<TaskStatus> => {
    setError(null);

    try {
      const status = await claudeService.getTaskStatus(requestId);
      setTaskStatus(status);
      return status;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to get task status';
      setError(errorMessage);
      throw err;
    }
  }, []);

  const cancelTask = useCallback(async (requestId: string): Promise<void> => {
    setError(null);

    try {
      await claudeService.cancelTask(requestId);
      
      // Update status if we're tracking this task
      if (taskStatus && taskStatus.request_id === requestId) {
        setTaskStatus({
          ...taskStatus,
          status: 'cancelled',
          updated_at: new Date().toISOString(),
        });
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to cancel task';
      setError(errorMessage);
      throw err;
    }
  }, [taskStatus]);

  const runTasksAndWait = useCallback(async (request: ReqRunTasksClaude): Promise<TaskStatus> => {
    setIsLoading(true);
    setIsPolling(true);
    setError(null);

    try {
      // First submit the task
      const submitResponse = await claudeService.runTasks(request);
      setCurrentTask(submitResponse);

      // Then poll for completion
      const finalStatus = await claudeService.pollTaskStatus(
        submitResponse.request_id,
        (status) => {
          setTaskStatus(status);
        }
      );

      setTaskStatus(finalStatus);
      return finalStatus;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to execute task';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
      setIsPolling(false);
    }
  }, []);

  return {
    // State
    isLoading,
    isPolling,
    currentTask,
    taskStatus,
    error,

    // Actions
    runTasks,
    runTasksAndWait,
    getTaskStatus,
    cancelTask,
    clearError,
    reset,
  };
}