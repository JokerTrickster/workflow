import { 
  GitHubRepository, 
  GitHubUser, 
  GitHubRateLimitResponse, 
  GitHubAPIError,
  Repository, 
  RepositoriesResponse,
  RepositoriesQuery,
  PaginationInfo 
} from '@/types/github'

// GitHub API configuration
const GITHUB_API_BASE_URL = 'https://api.github.com'
const DEFAULT_PER_PAGE = 30
const MAX_PER_PAGE = 100

export class GitHubAPIError extends Error {
  constructor(
    message: string,
    public status: number,
    public response?: GitHubAPIError
  ) {
    super(message)
    this.name = 'GitHubAPIError'
  }
}

export class GitHubService {
  private accessToken: string

  constructor(accessToken: string) {
    this.accessToken = accessToken
  }

  private async makeRequest<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${GITHUB_API_BASE_URL}${endpoint}`
    
    const response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.accessToken}`,
        'Accept': 'application/vnd.github+json',
        'X-GitHub-Api-Version': '2022-11-28',
        'User-Agent': 'AI-Git-Workbench/1.0',
        ...options.headers,
      },
    })

    if (!response.ok) {
      const errorData: GitHubAPIError = await response.json().catch(() => ({
        message: response.statusText || 'Unknown error',
      }))
      
      throw new GitHubAPIError(
        errorData.message || `GitHub API error: ${response.status}`,
        response.status,
        errorData
      )
    }

    return response.json()
  }

  private normalizeRepository(repo: GitHubRepository): Repository {
    return {
      id: repo.id,
      name: repo.name,
      fullName: repo.full_name,
      description: repo.description,
      isPrivate: repo.private,
      isFork: repo.fork,
      language: repo.language,
      starCount: repo.stargazers_count,
      forkCount: repo.forks_count,
      size: repo.size,
      lastPushedAt: repo.pushed_at,
      createdAt: repo.created_at,
      updatedAt: repo.updated_at,
      url: repo.html_url,
      cloneUrl: repo.clone_url,
      topics: repo.topics || [],
      hasIssues: repo.has_issues,
      isArchived: repo.archived,
      defaultBranch: repo.default_branch,
      openIssuesCount: repo.open_issues_count,
    }
  }

  private parseLinkHeader(linkHeader: string | null): { next?: string; last?: string } {
    if (!linkHeader) return {}
    
    const links: { next?: string; last?: string } = {}
    const parts = linkHeader.split(',')
    
    for (const part of parts) {
      const [url, rel] = part.split(';')
      const cleanUrl = url.trim().slice(1, -1) // Remove < >
      const cleanRel = rel.trim().replace(/rel="([^"]+)"/, '$1')
      
      if (cleanRel === 'next') links.next = cleanUrl
      if (cleanRel === 'last') links.last = cleanUrl
    }
    
    return links
  }

  private extractPageFromUrl(url: string): number {
    const match = url.match(/[&?]page=(\d+)/)
    return match ? parseInt(match[1], 10) : 1
  }

  // Get current user information
  async getCurrentUser(): Promise<GitHubUser> {
    return this.makeRequest<GitHubUser>('/user')
  }

  // Get rate limit information
  async getRateLimit(): Promise<GitHubRateLimitResponse> {
    return this.makeRequest<GitHubRateLimitResponse>('/rate_limit')
  }

  // Get user repositories with pagination
  async getRepositories(query: RepositoriesQuery = {}): Promise<RepositoriesResponse> {
    const {
      page = 1,
      per_page = DEFAULT_PER_PAGE,
      sort = 'updated',
      direction = 'desc',
      type = 'all'
    } = query

    // Ensure per_page doesn't exceed GitHub's limit
    const perPage = Math.min(per_page, MAX_PER_PAGE)
    
    const searchParams = new URLSearchParams({
      page: page.toString(),
      per_page: perPage.toString(),
      sort,
      direction,
      type
    })

    const endpoint = `/user/repos?${searchParams.toString()}`
    
    // Make the request and capture headers for pagination
    const url = `${GITHUB_API_BASE_URL}${endpoint}`
    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${this.accessToken}`,
        'Accept': 'application/vnd.github+json',
        'X-GitHub-Api-Version': '2022-11-28',
        'User-Agent': 'AI-Git-Workbench/1.0',
      },
    })

    if (!response.ok) {
      const errorData: GitHubAPIError = await response.json().catch(() => ({
        message: response.statusText || 'Unknown error',
      }))
      
      throw new GitHubAPIError(
        errorData.message || `GitHub API error: ${response.status}`,
        response.status,
        errorData
      )
    }

    const repositories: GitHubRepository[] = await response.json()
    
    // Parse pagination info from headers
    const linkHeader = response.headers.get('link')
    const links = this.parseLinkHeader(linkHeader)
    
    // Get rate limit info from headers
    const rateLimitRemaining = parseInt(response.headers.get('x-ratelimit-remaining') || '0', 10)
    const rateLimitReset = response.headers.get('x-ratelimit-reset') || '0'
    const resetAt = new Date(parseInt(rateLimitReset, 10) * 1000).toISOString()

    // Calculate total count (GitHub doesn't provide this in headers for /user/repos)
    // We'll estimate based on pagination
    let totalCount = repositories.length
    if (links.last) {
      const lastPage = this.extractPageFromUrl(links.last)
      totalCount = (lastPage - 1) * perPage + repositories.length
    } else if (repositories.length === perPage) {
      // If we got a full page, there might be more
      totalCount = page * perPage + 1
    } else {
      // Partial page means this is the last page
      totalCount = (page - 1) * perPage + repositories.length
    }

    return {
      repositories: repositories.map(repo => this.normalizeRepository(repo)),
      totalCount,
      hasNextPage: !!links.next,
      nextCursor: links.next ? page + 1 + '' : undefined,
      rateLimit: {
        remaining: rateLimitRemaining,
        resetAt
      }
    }
  }

  // Get all repositories (with automatic pagination)
  async getAllRepositories(
    query: Omit<RepositoriesQuery, 'page'> = {},
    maxPages: number = 10
  ): Promise<Repository[]> {
    const allRepositories: Repository[] = []
    let currentPage = 1
    let hasNextPage = true

    while (hasNextPage && currentPage <= maxPages) {
      const response = await this.getRepositories({
        ...query,
        page: currentPage,
        per_page: MAX_PER_PAGE // Use maximum page size for efficiency
      })

      allRepositories.push(...response.repositories)
      hasNextPage = response.hasNextPage
      currentPage++

      // Add a small delay to avoid hitting rate limits
      if (hasNextPage) {
        await new Promise(resolve => setTimeout(resolve, 100))
      }
    }

    return allRepositories
  }

  // Get a specific repository
  async getRepository(owner: string, repo: string): Promise<Repository> {
    const repository = await this.makeRequest<GitHubRepository>(`/repos/${owner}/${repo}`)
    return this.normalizeRepository(repository)
  }

  // Search repositories
  async searchRepositories(
    query: string, 
    options: {
      sort?: 'stars' | 'forks' | 'help-wanted-issues' | 'updated'
      order?: 'desc' | 'asc'
      per_page?: number
      page?: number
    } = {}
  ): Promise<{
    repositories: Repository[]
    totalCount: number
    incompleteResults: boolean
  }> {
    const {
      sort = 'updated',
      order = 'desc',
      per_page = DEFAULT_PER_PAGE,
      page = 1
    } = options

    const searchParams = new URLSearchParams({
      q: query,
      sort,
      order,
      per_page: Math.min(per_page, MAX_PER_PAGE).toString(),
      page: page.toString()
    })

    const response = await this.makeRequest<{
      total_count: number
      incomplete_results: boolean
      items: GitHubRepository[]
    }>(`/search/repositories?${searchParams.toString()}`)

    return {
      repositories: response.items.map(repo => this.normalizeRepository(repo)),
      totalCount: response.total_count,
      incompleteResults: response.incomplete_results
    }
  }
}

// Utility function to create GitHub service with access token
export function createGitHubService(accessToken: string): GitHubService {
  return new GitHubService(accessToken)
}

// Helper function to extract token from Supabase session
export function extractGitHubToken(session: any): string | null {
  return session?.provider_token || null
}