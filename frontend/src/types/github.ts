// GitHub API Response Types

export interface GitHubRepository {
  id: number
  name: string
  full_name: string
  description: string | null
  private: boolean
  fork: boolean
  homepage: string | null
  html_url: string
  clone_url: string
  language: string | null
  size: number
  stargazers_count: number
  watchers_count: number
  forks_count: number
  archived: boolean
  disabled: boolean
  pushed_at: string | null
  created_at: string
  updated_at: string
  default_branch: string
  topics: string[]
  has_issues: boolean
  has_projects: boolean
  has_wiki: boolean
  has_pages: boolean
  has_downloads: boolean
  visibility: 'public' | 'private' | 'internal'
  open_issues_count: number
}

export interface GitHubUser {
  id: number
  login: string
  avatar_url: string
  html_url: string
  name: string | null
  company: string | null
  blog: string | null
  location: string | null
  email: string | null
  bio: string | null
  public_repos: number
  public_gists: number
  followers: number
  following: number
  created_at: string
  updated_at: string
}

// Rate Limit Response
export interface GitHubRateLimit {
  limit: number
  used: number
  remaining: number
  reset: number
  resource: string
}

export interface GitHubRateLimitResponse {
  resources: {
    core: GitHubRateLimit
    search: GitHubRateLimit
    graphql: GitHubRateLimit
    integration_manifest: GitHubRateLimit
    source_import: GitHubRateLimit
    code_scanning_upload: GitHubRateLimit
    actions_runner_registration: GitHubRateLimit
    scim: GitHubRateLimit
    dependency_snapshots: GitHubRateLimit
  }
  rate: GitHubRateLimit
}

// Normalized types for our application
export interface Repository {
  id: number
  name: string
  fullName: string
  description: string | null
  isPrivate: boolean
  isFork: boolean
  language: string | null
  starCount: number
  forkCount: number
  size: number
  lastPushedAt: string | null
  createdAt: string
  updatedAt: string
  url: string
  cloneUrl: string
  topics: string[]
  hasIssues: boolean
  isArchived: boolean
  defaultBranch: string
  openIssuesCount: number
}

// API Response types
export interface RepositoriesResponse {
  repositories: Repository[]
  totalCount: number
  hasNextPage: boolean
  nextCursor?: string
  rateLimit: {
    remaining: number
    resetAt: string
  }
}

export interface GitHubAPIError {
  message: string
  documentation_url?: string
  errors?: Array<{
    resource: string
    field: string
    code: string
  }>
}

// Query parameters
export interface RepositoriesQuery {
  page?: number
  per_page?: number
  sort?: 'created' | 'updated' | 'pushed' | 'full_name'
  direction?: 'asc' | 'desc'
  type?: 'all' | 'owner' | 'public' | 'private' | 'member'
}

export interface PaginationInfo {
  page: number
  perPage: number
  totalCount: number
  totalPages: number
  hasNextPage: boolean
  hasPreviousPage: boolean
}