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

// Temporary generic types - Stream B will provide specific interfaces
interface GitHubIssue {
  id: number;
  number: number;
  title: string;
  body: string | null;
  state: 'open' | 'closed';
  created_at: string;
  updated_at: string;
  closed_at: string | null;
  html_url: string;
  user: {
    id: number;
    login: string;
    avatar_url: string;
  };
  labels: Array<{
    id: number;
    name: string;
    color: string;
  }>;
  assignees: Array<{
    id: number;
    login: string;
    avatar_url: string;
  }>;
}

interface GitHubPullRequest {
  id: number;
  number: number;
  title: string;
  body: string | null;
  state: 'open' | 'closed' | 'merged';
  created_at: string;
  updated_at: string;
  closed_at: string | null;
  merged_at: string | null;
  html_url: string;
  user: {
    id: number;
    login: string;
    avatar_url: string;
  };
  head: {
    ref: string;
    sha: string;
  };
  base: {
    ref: string;
    sha: string;
  };
  mergeable: boolean | null;
  draft: boolean;
}

interface GitHubIssuesResponse {
  issues: GitHubIssue[];
  nextCursor?: number;
  hasMore: boolean;
}

interface GitHubPullRequestsResponse {
  pullRequests: GitHubPullRequest[];
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

  private static async sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  private static async makeApiRequest(url: string, accessToken: string, retries = 3): Promise<Response> {
    for (let attempt = 1; attempt <= retries; attempt++) {
      try {
        const response = await fetch(url, {
          headers: {
            'Authorization': `Bearer ${accessToken}`,
            'Accept': 'application/vnd.github.v3+json',
            'User-Agent': 'AI-Git-Workbench/1.0.0'
          }
        });

        if (response.ok) {
          return response;
        }

        if (response.status === 401) {
          throw new Error('GitHub access token expired. Please sign in again.');
        }

        if (response.status === 403) {
          const rateLimitReset = response.headers.get('X-RateLimit-Reset');
          const resetTime = rateLimitReset 
            ? new Date(parseInt(rateLimitReset) * 1000).toLocaleTimeString() 
            : 'unknown';
          
          if (attempt < retries) {
            // Wait with exponential backoff for rate limit
            const waitTime = Math.pow(2, attempt) * 1000; // 2s, 4s, 8s
            console.warn(`Rate limit hit, waiting ${waitTime}ms before retry ${attempt}/${retries}`);
            await this.sleep(waitTime);
            continue;
          }
          
          throw new Error(`GitHub API rate limit exceeded. Resets at ${resetTime}.`);
        }

        if (response.status >= 500 && attempt < retries) {
          // Server error, retry with exponential backoff
          const waitTime = Math.pow(2, attempt) * 1000;
          console.warn(`Server error ${response.status}, waiting ${waitTime}ms before retry ${attempt}/${retries}`);
          await this.sleep(waitTime);
          continue;
        }

        throw new Error(`GitHub API error: ${response.status} ${response.statusText}`);
      } catch (error) {
        if (attempt === retries || error instanceof Error && error.message.includes('token expired')) {
          throw error;
        }
        
        // Network error, retry with exponential backoff
        const waitTime = Math.pow(2, attempt) * 500;
        console.warn(`Network error, waiting ${waitTime}ms before retry ${attempt}/${retries}:`, error);
        await this.sleep(waitTime);
      }
    }

    throw new Error('Max retries exceeded');
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
      const url = `https://api.github.com/user/repos?page=${page}&per_page=${perPage}&sort=updated&type=all`;
      const response = await this.makeApiRequest(url, accessToken);

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
      const url = 'https://api.github.com/user';
      const response = await this.makeApiRequest(url, accessToken);

      return await response.json();
    } catch (error) {
      console.error('Failed to fetch user profile from GitHub:', error);
      throw error;
    }
  }

  /**
   * Fetch issues for a specific repository
   * @param repoId - Repository ID (format: "owner/repo" e.g., "facebook/react")
   * @param page - Page number for pagination (default: 1)
   * @param perPage - Number of items per page (default: 30, max: 100)
   * @param state - Issue state filter: 'open', 'closed', or 'all' (default: 'open')
   */
  static async fetchRepositoryIssues(
    repoId: string, 
    page = 1, 
    perPage = 30,
    state: 'open' | 'closed' | 'all' = 'open'
  ): Promise<GitHubIssuesResponse> {
    const accessToken = await this.getAccessToken();
    
    if (!accessToken) {
      throw new Error('No GitHub access token available. Please sign in with GitHub.');
    }

    // Validate repoId format
    if (!repoId.includes('/')) {
      throw new Error('Repository ID must be in format "owner/repo"');
    }

    try {
      const url = `https://api.github.com/repos/${repoId}/issues?page=${page}&per_page=${perPage}&state=${state}&sort=updated`;
      const response = await this.makeApiRequest(url, accessToken);

      const issues: GitHubIssue[] = await response.json();
      
      // Filter out pull requests (GitHub API returns PRs as issues)
      const filteredIssues = issues.filter(issue => !issue.html_url.includes('/pull/'));

      // Check if there are more pages by looking at the Link header
      const linkHeader = response.headers.get('Link');
      const hasNext = linkHeader?.includes('rel="next"') || false;

      return {
        issues: filteredIssues,
        nextCursor: hasNext ? page + 1 : undefined,
        hasMore: hasNext
      };
    } catch (error) {
      console.error(`Failed to fetch issues for repository ${repoId}:`, error);
      throw error;
    }
  }

  /**
   * Fetch pull requests for a specific repository
   * @param repoId - Repository ID (format: "owner/repo" e.g., "facebook/react")
   * @param page - Page number for pagination (default: 1)
   * @param perPage - Number of items per page (default: 30, max: 100)
   * @param state - PR state filter: 'open', 'closed', or 'all' (default: 'open')
   */
  static async fetchRepositoryPullRequests(
    repoId: string, 
    page = 1, 
    perPage = 30,
    state: 'open' | 'closed' | 'all' = 'open'
  ): Promise<GitHubPullRequestsResponse> {
    const accessToken = await this.getAccessToken();
    
    if (!accessToken) {
      throw new Error('No GitHub access token available. Please sign in with GitHub.');
    }

    // Validate repoId format
    if (!repoId.includes('/')) {
      throw new Error('Repository ID must be in format "owner/repo"');
    }

    try {
      const url = `https://api.github.com/repos/${repoId}/pulls?page=${page}&per_page=${perPage}&state=${state}&sort=updated`;
      const response = await this.makeApiRequest(url, accessToken);

      const pullRequests: GitHubPullRequest[] = await response.json();

      // Check if there are more pages by looking at the Link header
      const linkHeader = response.headers.get('Link');
      const hasNext = linkHeader?.includes('rel="next"') || false;

      return {
        pullRequests,
        nextCursor: hasNext ? page + 1 : undefined,
        hasMore: hasNext
      };
    } catch (error) {
      console.error(`Failed to fetch pull requests for repository ${repoId}:`, error);
      throw error;
    }
  }
}