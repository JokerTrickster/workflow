'use client';

import { useQuery } from '@tanstack/react-query';
import { GitHubApiService } from '../services/githubApi';

interface GitHubEvent {
  id: string;
  type: string;
  actor: {
    login: string;
    avatar_url: string;
  };
  repo: {
    name: string;
  };
  payload: any;
  created_at: string;
}

interface UseGitHubEventsParams {
  repoId: string; // Format: "owner/repo"
  enabled?: boolean;
}

export function useGitHubEvents({ repoId, enabled = true }: UseGitHubEventsParams) {
  const queryKey = ['github-events', repoId];

  return useQuery({
    queryKey,
    queryFn: async (): Promise<GitHubEvent[]> => {
      if (!repoId || !repoId.includes('/')) {
        throw new Error('Invalid repository ID format. Expected "owner/repo"');
      }
      
      // Use existing API service pattern
      const accessToken = await (GitHubApiService as any).getAccessToken();
      if (!accessToken) {
        throw new Error('No GitHub access token available. Please sign in with GitHub.');
      }

      const url = `https://api.github.com/repos/${repoId}/events?per_page=30`;
      const response = await (GitHubApiService as any).makeApiRequest(url, accessToken);
      
      return await response.json();
    },
    enabled: enabled && !!repoId,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      // Don't retry on auth errors
      if (error instanceof Error && error.message.includes('token expired')) {
        return false;
      }
      return failureCount < 2;
    },
  });
}