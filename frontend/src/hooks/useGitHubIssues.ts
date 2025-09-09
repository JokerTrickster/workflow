'use client';

import { useQuery } from '@tanstack/react-query';
import { GitHubApiService } from '../services/githubApi';
import { GitHubIssuesResponse, GitHubIssuesRequestParams } from '../types/github';

interface UseGitHubIssuesParams {
  repoId: string; // Format: "owner/repo"
  params?: Partial<GitHubIssuesRequestParams>;
  enabled?: boolean;
}

export function useGitHubIssues({ repoId, params = {}, enabled = true }: UseGitHubIssuesParams) {
  const queryKey = ['github-issues', repoId, params];

  return useQuery({
    queryKey,
    queryFn: async (): Promise<GitHubIssuesResponse> => {
      if (!repoId || !repoId.includes('/')) {
        throw new Error('Invalid repository ID format. Expected "owner/repo"');
      }
      
      return await GitHubApiService.fetchRepositoryIssues(repoId, params);
    },
    enabled: enabled && !!repoId,
    staleTime: 2 * 60 * 1000, // 2 minutes
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