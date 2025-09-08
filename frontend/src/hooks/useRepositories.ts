'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Repository } from '../domain/entities/Repository';
import { GitHubApiService } from '../services/githubApi';
import { ActivityLogger } from '../services/ActivityLogger';
import { useCallback, useEffect, useState } from 'react';

const REPOSITORIES_QUERY_KEY = 'repositories';
const CONNECTION_STORAGE_KEY = 'repository_connections';

interface RepositoryConnectionState {
  [repoId: number]: {
    is_connected: boolean;
    local_path?: string;
    connected_at?: string;
  };
}

interface UseRepositoriesReturn {
  repositories: Repository[];
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
  connectRepository: (repoId: number, localPath?: string) => Promise<void>;
  disconnectRepository: (repoId: number) => Promise<void>;
  isConnecting: boolean;
  isDisconnecting: boolean;
  getConnectionStatus: (repoId: number) => {
    is_connected: boolean;
    local_path?: string;
    connected_at?: string;
  };
}

export function useRepositories(): UseRepositoriesReturn {
  const queryClient = useQueryClient();
  const [connectionStates, setConnectionStates] = useState<RepositoryConnectionState>({});
  const activityLogger = ActivityLogger.getInstance();

  // Load connection states from localStorage on mount
  useEffect(() => {
    if (typeof window !== 'undefined') {
      try {
        const stored = localStorage.getItem(CONNECTION_STORAGE_KEY);
        if (stored) {
          setConnectionStates(JSON.parse(stored));
        }
      } catch (error) {
        console.error('Failed to load repository connections from localStorage:', error);
      }
    }
  }, []);

  // Save connection states to localStorage whenever they change
  const saveConnectionStates = useCallback((states: RepositoryConnectionState) => {
    setConnectionStates(states);
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem(CONNECTION_STORAGE_KEY, JSON.stringify(states));
      } catch (error) {
        console.error('Failed to save repository connections to localStorage:', error);
      }
    }
  }, []);

  // Fetch repositories from GitHub API
  const {
    data: repositoriesData,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: [REPOSITORIES_QUERY_KEY],
    queryFn: async () => {
      const response = await GitHubApiService.fetchUserRepositories();
      return response.repositories;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
  });

  // Merge repository data with connection states
  const repositories: Repository[] = (repositoriesData || []).map(repo => ({
    ...repo,
    is_connected: connectionStates[repo.id]?.is_connected || false,
    local_path: connectionStates[repo.id]?.local_path,
  }));

  // Connect repository mutation
  const connectMutation = useMutation({
    mutationFn: async ({ repoId, localPath }: { repoId: number; localPath?: string }) => {
      
      try {
        // Here you could add API call to backend if needed
        // For now, we'll just update local state
        const newState = {
          ...connectionStates,
          [repoId]: {
            is_connected: true,
            local_path: localPath,
            connected_at: new Date().toISOString(),
          },
        };
        saveConnectionStates(newState);
        
        // Log successful connection
        const repo = repositories.find(r => r.id === repoId);
        if (repo) {
          activityLogger.logRepositoryConnected(repoId, repo.name, localPath);
        }
        
        return newState;
      } catch (error) {
        // Log failed connection
        const repo = repositories.find(r => r.id === repoId);
        if (repo) {
          activityLogger.logRepositoryConnectionFailed(repo.name, error instanceof Error ? error.message : 'Unknown error');
        }
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate and refetch repositories to update UI
      queryClient.invalidateQueries({ queryKey: [REPOSITORIES_QUERY_KEY] });
    },
    onError: (error) => {
      console.error('Failed to connect repository:', error);
    },
  });

  // Disconnect repository mutation
  const disconnectMutation = useMutation({
    mutationFn: async (repoId: number) => {
      // Log disconnection
      const repo = repositories.find(r => r.id === repoId);
      
      // Here you could add API call to backend if needed
      // For now, we'll just update local state
      const newState = {
        ...connectionStates,
        [repoId]: {
          ...connectionStates[repoId],
          is_connected: false,
          local_path: undefined,
        },
      };
      saveConnectionStates(newState);
      
      if (repo) {
        activityLogger.logRepositoryDisconnected(repoId, repo.name);
      }
      
      return newState;
    },
    onSuccess: () => {
      // Invalidate and refetch repositories to update UI
      queryClient.invalidateQueries({ queryKey: [REPOSITORIES_QUERY_KEY] });
    },
    onError: (error) => {
      console.error('Failed to disconnect repository:', error);
    },
  });

  // Connect repository function
  const connectRepository = useCallback(async (repoId: number, localPath?: string) => {
    await connectMutation.mutateAsync({ repoId, localPath });
  }, [connectMutation]);

  // Disconnect repository function
  const disconnectRepository = useCallback(async (repoId: number) => {
    await disconnectMutation.mutateAsync(repoId);
  }, [disconnectMutation]);

  // Get connection status for a specific repository
  const getConnectionStatus = useCallback((repoId: number) => {
    return connectionStates[repoId] || {
      is_connected: false,
    };
  }, [connectionStates]);

  return {
    repositories,
    isLoading,
    isError,
    error: error as Error | null,
    refetch,
    connectRepository,
    disconnectRepository,
    isConnecting: connectMutation.isPending,
    isDisconnecting: disconnectMutation.isPending,
    getConnectionStatus,
  };
}