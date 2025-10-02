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

      try {
        return await GitHubApiService.fetchRepositoryIssues(repoId, params);
      } catch (error) {
        console.error('GitHub Issues API error:', error);
        // Return empty response instead of throwing to prevent TaskTab crash
        return {
          issues: [],
          hasMore: false
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