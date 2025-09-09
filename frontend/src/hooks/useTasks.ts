'use client';

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Task } from '../domain/entities/Task';
import { ApiDataSourceImpl } from '../data/datasources/ApiDataSourceImpl';

interface UseTasksParams {
  repositoryId: number;
  enabled?: boolean;
}

export function useTasks({ repositoryId, enabled = true }: UseTasksParams) {
  const queryClient = useQueryClient();
  const apiDataSource = new ApiDataSourceImpl();

  const queryKey = ['tasks', repositoryId];

  const query = useQuery({
    queryKey,
    queryFn: async (): Promise<Task[]> => {
      return await apiDataSource.getTasks(repositoryId);
    },
    enabled: enabled && !!repositoryId,
    staleTime: 30 * 1000, // 30 seconds
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      // Don't retry on auth errors
      if (error instanceof Error && (
        error.message.includes('unauthorized') ||
        error.message.includes('forbidden')
      )) {
        return false;
      }
      return failureCount < 2;
    },
  });

  const refetch = () => {
    queryClient.invalidateQueries({ queryKey });
  };

  return {
    ...query,
    refetch
  };
}