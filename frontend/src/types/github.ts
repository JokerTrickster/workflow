// GitHub API types for Issues and Pull Requests
// Based on GitHub REST API v3 response structures

// Base GitHub User interface
export interface GitHubUser {
  id: number;
  login: string;
  node_id: string;
  avatar_url: string;
  gravatar_id: string | null;
  url: string;
  html_url: string;
  followers_url: string;
  following_url: string;
  gists_url: string;
  starred_url: string;
  subscriptions_url: string;
  organizations_url: string;
  repos_url: string;
  events_url: string;
  received_events_url: string;
  type: 'User' | 'Bot' | 'Organization';
  site_admin: boolean;
}

// GitHub Label interface
export interface GitHubLabel {
  id: number;
  node_id: string;
  url: string;
  name: string;
  color: string;
  default: boolean;
  description: string | null;
}

// GitHub Milestone interface
export interface GitHubMilestone {
  id: number;
  node_id: string;
  url: string;
  html_url: string;
  labels_url: string;
  number: number;
  title: string;
  description: string | null;
  creator: GitHubUser;
  open_issues: number;
  closed_issues: number;
  state: 'open' | 'closed';
  created_at: string;
  updated_at: string;
  due_on: string | null;
  closed_at: string | null;
}

// GitHub Reactions interface
export interface GitHubReactions {
  url: string;
  total_count: number;
  '+1': number;
  '-1': number;
  laugh: number;
  hooray: number;
  confused: number;
  heart: number;
  rocket: number;
  eyes: number;
}

// GitHub Issue interface
export interface GitHubIssue {
  id: number;
  node_id: string;
  url: string;
  repository_url: string;
  labels_url: string;
  comments_url: string;
  events_url: string;
  html_url: string;
  number: number;
  title: string;
  body: string | null;
  state: 'open' | 'closed';
  state_reason?: 'completed' | 'not_planned' | 'reopened' | null;
  locked: boolean;
  active_lock_reason?: 'resolved' | 'off-topic' | 'too heated' | 'spam' | null;
  user: GitHubUser;
  labels: GitHubLabel[];
  assignee: GitHubUser | null;
  assignees: GitHubUser[];
  milestone: GitHubMilestone | null;
  comments: number;
  created_at: string;
  updated_at: string;
  closed_at: string | null;
  author_association: 'OWNER' | 'COLLABORATOR' | 'MEMBER' | 'CONTRIBUTOR' | 'FIRST_TIME_CONTRIBUTOR' | 'FIRST_TIMER' | 'NONE';
  reactions?: GitHubReactions;
  timeline_url?: string;
  performed_via_github_app?: {
    id: number;
    slug: string;
    node_id: string;
    owner: GitHubUser;
    name: string;
    description: string;
    external_url: string;
    html_url: string;
    created_at: string;
    updated_at: string;
    permissions: {
      metadata: string;
      contents: string;
      issues: string;
      single_file: string;
    };
    events: string[];
  } | null;
  draft?: boolean;
}

// GitHub Git Reference interface
export interface GitHubGitRef {
  label: string;
  ref: string;
  sha: string;
  user: GitHubUser;
  repo: GitHubRepository;
}

// Simplified GitHub Repository interface for PR references
export interface GitHubRepository {
  id: number;
  node_id: string;
  name: string;
  full_name: string;
  owner: GitHubUser;
  private: boolean;
  html_url: string;
  description: string | null;
  fork: boolean;
  url: string;
  created_at: string;
  updated_at: string;
  pushed_at: string;
  clone_url: string;
  size: number;
  stargazers_count: number;
  watchers_count: number;
  language: string | null;
  forks_count: number;
  archived: boolean;
  disabled: boolean;
  open_issues_count: number;
  license: {
    key: string;
    name: string;
    spdx_id: string;
    url: string;
    node_id: string;
  } | null;
  allow_forking: boolean;
  is_template: boolean;
  topics: string[];
  visibility: 'public' | 'private' | 'internal';
  forks: number;
  open_issues: number;
  watchers: number;
  default_branch: string;
}

// GitHub Pull Request interface
export interface GitHubPullRequest {
  id: number;
  node_id: string;
  url: string;
  html_url: string;
  diff_url: string;
  patch_url: string;
  issue_url: string;
  commits_url: string;
  review_comments_url: string;
  review_comment_url: string;
  comments_url: string;
  statuses_url: string;
  number: number;
  title: string;
  body: string | null;
  state: 'open' | 'closed';
  locked: boolean;
  active_lock_reason?: 'resolved' | 'off-topic' | 'too heated' | 'spam' | null;
  user: GitHubUser;
  labels: GitHubLabel[];
  milestone: GitHubMilestone | null;
  assignee: GitHubUser | null;
  assignees: GitHubUser[];
  requested_reviewers: GitHubUser[];
  requested_teams: Array<{
    id: number;
    node_id: string;
    url: string;
    html_url: string;
    name: string;
    slug: string;
    description: string | null;
    privacy: 'closed' | 'secret';
    permission: string;
    members_url: string;
    repositories_url: string;
    parent: {
      id: number;
      node_id: string;
      url: string;
      html_url: string;
      name: string;
      slug: string;
      description: string | null;
      privacy: 'closed' | 'secret';
      permission: string;
    } | null;
  }>;
  head: GitHubGitRef;
  base: GitHubGitRef;
  _links: {
    self: { href: string };
    html: { href: string };
    issue: { href: string };
    comments: { href: string };
    review_comments: { href: string };
    review_comment: { href: string };
    commits: { href: string };
    statuses: { href: string };
  };
  author_association: 'OWNER' | 'COLLABORATOR' | 'MEMBER' | 'CONTRIBUTOR' | 'FIRST_TIME_CONTRIBUTOR' | 'FIRST_TIMER' | 'NONE';
  auto_merge: {
    enabled_by: GitHubUser;
    merge_method: 'merge' | 'squash' | 'rebase';
    commit_title: string;
    commit_message: string;
  } | null;
  draft: boolean;
  merged: boolean;
  mergeable: boolean | null;
  rebaseable?: boolean | null;
  mergeable_state: 'clean' | 'dirty' | 'unstable' | 'blocked' | 'behind' | 'draft' | 'has_hooks' | 'unknown';
  merged_by: GitHubUser | null;
  comments: number;
  review_comments: number;
  maintainer_can_modify: boolean;
  commits: number;
  additions: number;
  deletions: number;
  changed_files: number;
  created_at: string;
  updated_at: string;
  closed_at: string | null;
  merged_at: string | null;
  merge_commit_sha: string | null;
  reactions?: GitHubReactions;
}

// Pagination interfaces
export interface GitHubPaginationLinks {
  first?: string;
  prev?: string;
  next?: string;
  last?: string;
}

export interface GitHubPaginationMeta {
  page: number;
  per_page: number;
  total_count?: number;
  has_next_page: boolean;
  has_previous_page: boolean;
}

// API Response interfaces for Issues
export interface GitHubIssuesResponse {
  issues: GitHubIssue[];
  nextCursor?: number;
  hasMore: boolean;
  pagination?: GitHubPaginationMeta;
  links?: GitHubPaginationLinks;
}

// API Response interfaces for Pull Requests
export interface GitHubPullRequestsResponse {
  pullRequests: GitHubPullRequest[];
  nextCursor?: number;
  hasMore: boolean;
  pagination?: GitHubPaginationMeta;
  links?: GitHubPaginationLinks;
}

// Request parameter interfaces
export interface GitHubIssuesRequestParams {
  milestone?: string | number | 'none' | '*';
  state?: 'open' | 'closed' | 'all';
  assignee?: string | 'none' | '*';
  creator?: string;
  mentioned?: string;
  labels?: string | string[];
  sort?: 'created' | 'updated' | 'comments';
  direction?: 'asc' | 'desc';
  since?: string; // ISO 8601 timestamp
  per_page?: number; // 1-100
  page?: number;
}

export interface GitHubPullRequestsRequestParams {
  state?: 'open' | 'closed' | 'all';
  head?: string; // Format: user:ref-name or organization:ref-name
  base?: string; // Base branch name
  sort?: 'created' | 'updated' | 'popularity' | 'long-running';
  direction?: 'asc' | 'desc';
  per_page?: number; // 1-100
  page?: number;
}

// Error interfaces
export interface GitHubApiError {
  message: string;
  documentation_url?: string;
  errors?: Array<{
    resource: string;
    field: string;
    code: string;
  }>;
}

// Rate limit interface
export interface GitHubRateLimit {
  limit: number;
  remaining: number;
  reset: number; // Unix timestamp
  used: number;
  resource: 'core' | 'search' | 'graphql' | 'integration_manifest' | 'source_import' | 'code_scanning_upload';
}

// Search interfaces for future extensibility
export interface GitHubSearchIssuesParams {
  q: string; // Search query
  sort?: 'comments' | 'reactions' | 'reactions-+1' | 'reactions--1' | 'reactions-smile' | 'reactions-thinking_face' | 'reactions-heart' | 'reactions-tada' | 'interactions' | 'created' | 'updated';
  order?: 'asc' | 'desc';
  per_page?: number; // 1-100
  page?: number;
}

export interface GitHubSearchIssuesResponse {
  total_count: number;
  incomplete_results: boolean;
  items: GitHubIssue[];
}

// Type guards for runtime type checking
export function isGitHubIssue(item: unknown): item is GitHubIssue {
  return typeof item === 'object' && 
         item !== null && 
         typeof item.id === 'number' && 
         typeof item.number === 'number' && 
         typeof item.title === 'string' && 
         typeof item.state === 'string' &&
         !item.pull_request; // Issues don't have pull_request property
}

export function isGitHubPullRequest(item: unknown): item is GitHubPullRequest {
  return typeof item === 'object' && 
         item !== null && 
         typeof item.id === 'number' && 
         typeof item.number === 'number' && 
         typeof item.title === 'string' && 
         typeof item.state === 'string' &&
         typeof item.head === 'object' &&
         typeof item.base === 'object';
}

// Repository-related types (simplified for this context)
export interface GitHubRepo {
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

export interface GitHubApiResponse<T = unknown> {
  repositories: T[];
  nextCursor?: number;
  hasMore: boolean;
}

// Export all commonly used types for convenience
export type {
  GitHubUser,
  GitHubLabel,
  GitHubMilestone,
  GitHubReactions,
  GitHubIssue,
  GitHubPullRequest,
  GitHubRepository,
  GitHubGitRef,
  GitHubIssuesResponse,
  GitHubPullRequestsResponse,
  GitHubIssuesRequestParams,
  GitHubPullRequestsRequestParams,
  GitHubApiError,
  GitHubRateLimit,
  GitHubPaginationMeta,
  GitHubPaginationLinks,
  GitHubSearchIssuesParams,
  GitHubSearchIssuesResponse
};