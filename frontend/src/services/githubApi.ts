import { supabase } from '@/lib/supabase/client';
import { Repository } from '@/domain/entities/Repository';
import {
  GitHubRepo,
  GitHubApiResponse,
  GitHubIssue,
  GitHubPullRequest,
  GitHubIssuesResponse,
  GitHubPullRequestsResponse,
  GitHubIssuesRequestParams,
  GitHubPullRequestsRequestParams
} from '@/types/github';

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

  private static parseLinkHeader(linkHeader: string): { first?: string; prev?: string; next?: string; last?: string } {
    const links: { first?: string; prev?: string; next?: string; last?: string } = {};
    
    // Parse Link header format: <url1>; rel="next", <url2>; rel="last"
    const parts = linkHeader.split(',');
    for (const part of parts) {
      const match = part.match(/<([^>]+)>;\s*rel="([^"]+)"/);
      if (match) {
        const [, url, rel] = match;
        if (rel === 'first' || rel === 'prev' || rel === 'next' || rel === 'last') {
          links[rel] = url;
        }
      }
    }
    
    return links;
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
   * @param params - Optional parameters for filtering and pagination
   */
  static async fetchRepositoryIssues(
    repoId: string, 
    params: Partial<GitHubIssuesRequestParams> = {}
  ): Promise<GitHubIssuesResponse> {
    // Set defaults
    const {
      page = 1,
      per_page = 30,
      state = 'open',
      sort = 'updated',
      direction = 'desc',
      ...otherParams
    } = params;
    const accessToken = await this.getAccessToken();
    
    if (!accessToken) {
      throw new Error('No GitHub access token available. Please sign in with GitHub.');
    }

    // Validate repoId format
    if (!repoId.includes('/')) {
      throw new Error('Repository ID must be in format "owner/repo"');
    }

    try {
      // Build query parameters
      const queryParams = new URLSearchParams({
        page: page.toString(),
        per_page: per_page.toString(),
        state,
        sort,
        direction,
        ...Object.fromEntries(
          Object.entries(otherParams).map(([key, value]) => [
            key, 
            Array.isArray(value) ? value.join(',') : String(value)
          ])
        )
      });

      const url = `https://api.github.com/repos/${repoId}/issues?${queryParams}`;
      const response = await this.makeApiRequest(url, accessToken);

      const issues: GitHubIssue[] = await response.json();
      
      // Filter out pull requests (GitHub API returns PRs as issues)
      const filteredIssues = issues.filter(issue => !issue.html_url.includes('/pull/'));

      // Parse pagination metadata from headers
      const linkHeader = response.headers.get('Link');
      const hasNext = linkHeader?.includes('rel="next"') || false;
      const rateLimitRemaining = response.headers.get('X-RateLimit-Remaining');
      const rateLimitReset = response.headers.get('X-RateLimit-Reset');

      return {
        issues: filteredIssues,
        nextCursor: hasNext ? page + 1 : undefined,
        hasMore: hasNext,
        pagination: {
          page,
          per_page,
          has_next_page: hasNext,
          has_previous_page: page > 1
        },
        links: linkHeader ? this.parseLinkHeader(linkHeader) : undefined
      };
    } catch (error) {
      console.error(`Failed to fetch issues for repository ${repoId}:`, error);
      throw error;
    }
  }

  /**
   * Fetch pull requests for a specific repository
   * @param repoId - Repository ID (format: "owner/repo" e.g., "facebook/react")
   * @param params - Optional parameters for filtering and pagination
   */
  static async fetchRepositoryPullRequests(
    repoId: string, 
    params: Partial<GitHubPullRequestsRequestParams> = {}
  ): Promise<GitHubPullRequestsResponse> {
    // Set defaults
    const {
      page = 1,
      per_page = 30,
      state = 'open',
      sort = 'updated',
      direction = 'desc',
      ...otherParams
    } = params;
    const accessToken = await this.getAccessToken();
    
    if (!accessToken) {
      throw new Error('No GitHub access token available. Please sign in with GitHub.');
    }

    // Validate repoId format
    if (!repoId.includes('/')) {
      throw new Error('Repository ID must be in format "owner/repo"');
    }

    try {
      // Build query parameters
      const queryParams = new URLSearchParams({
        page: page.toString(),
        per_page: per_page.toString(),
        state,
        sort,
        direction,
        ...Object.fromEntries(
          Object.entries(otherParams).map(([key, value]) => [
            key, 
            Array.isArray(value) ? value.join(',') : String(value)
          ])
        )
      });

      const url = `https://api.github.com/repos/${repoId}/pulls?${queryParams}`;
      const response = await this.makeApiRequest(url, accessToken);

      const pullRequests: GitHubPullRequest[] = await response.json();

      // Parse pagination metadata from headers
      const linkHeader = response.headers.get('Link');
      const hasNext = linkHeader?.includes('rel="next"') || false;
      const rateLimitRemaining = response.headers.get('X-RateLimit-Remaining');
      const rateLimitReset = response.headers.get('X-RateLimit-Reset');

      return {
        pullRequests,
        nextCursor: hasNext ? page + 1 : undefined,
        hasMore: hasNext,
        pagination: {
          page,
          per_page,
          has_next_page: hasNext,
          has_previous_page: page > 1
        },
        links: linkHeader ? this.parseLinkHeader(linkHeader) : undefined
      };
    } catch (error) {
      console.error(`Failed to fetch pull requests for repository ${repoId}:`, error);
      throw error;
    }
  }
}