import { supabase } from '@/lib/supabase/client';
import { Repository } from '@/domain/entities/Repository';

interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  description: string | null;
  html_url: string;
  clone_url: string;
  ssh_url: string;
  private: boolean;
  language: string | null;
  stargazers_count: number;
  forks_count: number;
  created_at: string;
  updated_at: string;
  pushed_at: string;
  default_branch: string;
}

interface GitHubApiResponse {
  repositories: Repository[];
  nextCursor?: number;
  hasMore: boolean;
}

export class GitHubApiService {
  private static async getAccessToken(): Promise<string | null> {
    const { data: { session }, error } = await supabase.auth.getSession();
    
    if (error) {
      console.error('Error getting session:', error);
      return null;
    }

    if (!session?.provider_token) {
      console.error('No GitHub access token found in session');
      return null;
    }

    return session.provider_token;
  }

  private static transformGitHubRepo(repo: GitHubRepo): Repository {
    return {
      id: repo.id,
      name: repo.name,
      full_name: repo.full_name,
      description: repo.description || undefined,
      html_url: repo.html_url,
      clone_url: repo.clone_url,
      ssh_url: repo.ssh_url,
      private: repo.private,
      language: repo.language || undefined,
      stargazers_count: repo.stargazers_count,
      forks_count: repo.forks_count,
      created_at: repo.created_at,
      updated_at: repo.updated_at,
      pushed_at: repo.pushed_at,
      default_branch: repo.default_branch,
      is_connected: false, // Default to false, will be managed separately
    };
  }

  static async fetchUserRepositories(page = 1, perPage = 30): Promise<GitHubApiResponse> {
    const accessToken = await this.getAccessToken();
    
    if (!accessToken) {
      throw new Error('No GitHub access token available. Please sign in with GitHub.');
    }

    try {
      const response = await fetch(
        `https://api.github.com/user/repos?page=${page}&per_page=${perPage}&sort=updated&type=all`,
        {
          headers: {
            'Authorization': `Bearer ${accessToken}`,
            'Accept': 'application/vnd.github.v3+json',
            'User-Agent': 'AI-Git-Workbench/1.0.0'
          }
        }
      );

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('GitHub access token expired. Please sign in again.');
        }
        if (response.status === 403) {
          const rateLimitReset = response.headers.get('X-RateLimit-Reset');
          const resetTime = rateLimitReset 
            ? new Date(parseInt(rateLimitReset) * 1000).toLocaleTimeString() 
            : 'unknown';
          throw new Error(`GitHub API rate limit exceeded. Resets at ${resetTime}.`);
        }
        throw new Error(`GitHub API error: ${response.status} ${response.statusText}`);
      }

      const repos: GitHubRepo[] = await response.json();
      const transformedRepos = repos.map(this.transformGitHubRepo);

      // Check if there are more pages by looking at the Link header
      const linkHeader = response.headers.get('Link');
      const hasNext = linkHeader?.includes('rel="next"') || false;

      return {
        repositories: transformedRepos,
        nextCursor: hasNext ? page + 1 : undefined,
        hasMore: hasNext
      };
    } catch (error) {
      console.error('Failed to fetch repositories from GitHub:', error);
      throw error;
    }
  }

  static async fetchUserProfile() {
    const accessToken = await this.getAccessToken();
    
    if (!accessToken) {
      throw new Error('No GitHub access token available. Please sign in with GitHub.');
    }

    try {
      const response = await fetch('https://api.github.com/user', {
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Accept': 'application/vnd.github.v3+json',
          'User-Agent': 'AI-Git-Workbench/1.0.0'
        }
      });

      if (!response.ok) {
        throw new Error(`GitHub API error: ${response.status} ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Failed to fetch user profile from GitHub:', error);
      throw error;
    }
  }
}