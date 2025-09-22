import { useState, useCallback, useEffect } from 'react';
import {
  claudeService,
  TaskHistoryResponse,
  WorkflowHistory,
  PaginationMeta,
  TaskHistoryParams
} from '../services/claudeService';

export interface UseTaskHistoryResult {
  // State
  tasks: WorkflowHistory[];
  pagination: PaginationMeta | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  loadTaskHistory: (repositoryName: string, params?: TaskHistoryParams) => Promise<void>;
  loadNextPage: () => Promise<void>;
  loadPrevPage: () => Promise<void>;
  loadPage: (page: number) => Promise<void>;
  refresh: () => Promise<void>;
  clearError: () => void;
  reset: () => void;

  // Computed state
  hasNextPage: boolean;
  hasPrevPage: boolean;
  currentPage: number;
  totalPages: number;
}

interface TaskHistoryState {
  repositoryName: string | null;
  params: TaskHistoryParams;
}

/**
 * React hook for managing task history state and pagination
 */
export function useTaskHistory(): UseTaskHistoryResult {
  const [tasks, setTasks] = useState<WorkflowHistory[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [state, setState] = useState<TaskHistoryState>({
    repositoryName: null,
    params: { page: 1, limit: 20 }
  });

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const reset = useCallback(() => {
    setTasks([]);
    setPagination(null);
    setIsLoading(false);
    setError(null);
    setState({
      repositoryName: null,
      params: { page: 1, limit: 20 }
    });
  }, []);

  const loadTaskHistory = useCallback(async (
    repositoryName: string,
    params: TaskHistoryParams = {}
  ): Promise<void> => {
    if (!repositoryName.trim()) {
      setError('Repository name is required');
      return;
    }

    setIsLoading(true);
    setError(null);

    const mergedParams = {
      page: 1,
      limit: 20,
      ...params
    };

    try {
      const response = await claudeService.getTaskHistory(repositoryName, mergedParams);
      setTasks(response.data);
      setPagination(response.pagination);
      setState({
        repositoryName,
        params: mergedParams
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load task history';
      setError(errorMessage);
      setTasks([]);
      setPagination(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const loadPage = useCallback(async (page: number): Promise<void> => {
    if (!state.repositoryName) {
      setError('No repository selected');
      return;
    }

    if (page < 1) {
      setError('Invalid page number');
      return;
    }

    const newParams = { ...state.params, page };
    await loadTaskHistory(state.repositoryName, newParams);
  }, [state.repositoryName, state.params, loadTaskHistory]);

  const loadNextPage = useCallback(async (): Promise<void> => {
    if (!pagination || pagination.page >= pagination.total_pages) {
      return;
    }
    await loadPage(pagination.page + 1);
  }, [pagination, loadPage]);

  const loadPrevPage = useCallback(async (): Promise<void> => {
    if (!pagination || pagination.page <= 1) {
      return;
    }
    await loadPage(pagination.page - 1);
  }, [pagination, loadPage]);

  const refresh = useCallback(async (): Promise<void> => {
    if (!state.repositoryName) {
      return;
    }
    await loadTaskHistory(state.repositoryName, state.params);
  }, [state.repositoryName, state.params, loadTaskHistory]);

  // Computed values
  const hasNextPage = pagination ? pagination.page < pagination.total_pages : false;
  const hasPrevPage = pagination ? pagination.page > 1 : false;
  const currentPage = pagination?.page || 1;
  const totalPages = pagination?.total_pages || 0;

  return {
    // State
    tasks,
    pagination,
    isLoading,
    error,

    // Actions
    loadTaskHistory,
    loadNextPage,
    loadPrevPage,
    loadPage,
    refresh,
    clearError,
    reset,

    // Computed state
    hasNextPage,
    hasPrevPage,
    currentPage,
    totalPages,
  };
}