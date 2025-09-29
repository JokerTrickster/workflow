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

      try {
        return await GitHubApiService.fetchRepositoryPullRequests(repoId, params);
      } catch (error) {
        console.error('GitHub Pull Requests API error:', error);
        // Return empty response instead of throwing to prevent TaskTab crash
        return {
          pullRequests: [],
          total_count: 0,
          has_more: false
        };
      }
    },
    enabled: enabled && !!repoId,
    staleTime: 10 * 1000, // 10 seconds - very short for immediate updates
    refetchOnWindowFocus: false,
    refetchOnMount: true, // Always refetch when component mounts
    refetchInterval: false, // No automatic refetch interval
    gcTime: 5 * 60 * 1000, // 5 minutes garbage collection
    retry: (failureCount, error) => {
      // Don't retry on auth errors or API errors
      if (error instanceof Error && (
        error.message.includes('token expired') ||
        error.message.includes('authentication') ||
        error.message.includes('unauthorized') ||
        error.message.includes('forbidden')
      )) {
        return false;
      }
      return failureCount < 2;
    },
  });
}