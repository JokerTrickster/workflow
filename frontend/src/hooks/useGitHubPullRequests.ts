'use client';

import { useQuery } from '@tanstack/react-query';
import { GitHubApiService } from '../services/githubApi';
import { GitHubPullRequestsResponse, GitHubPullRequestsRequestParams } from '../types/github';

interface UseGitHubPullRequestsParams {
  repoId: string; // Format: "owner/repo"
  params?: Partial<GitHubPullRequestsRequestParams>;
  enabled?: boolean;
}

export function useGitHubPullRequests({ repoId, params = {}, enabled = true }: UseGitHubPullRequestsParams) {
  const queryKey = ['github-pull-requests', repoId, params];

  return useQuery({
    queryKey,
    queryFn: async (): Promise<GitHubPullRequestsResponse> => {
      if (!repoId || !repoId.includes('/')) {
        throw new Error('Invalid repository ID format. Expected "owner/repo"');
      }
      
      return await GitHubApiService.fetchRepositoryPullRequests(repoId, params);
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