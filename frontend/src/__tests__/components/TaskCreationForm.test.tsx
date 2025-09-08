import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { TaskCreationForm } from '../../components/TaskCreationForm';
import { GitHubIssue, GitHubPullRequest } from '../../types/github';

// Mock the UI components
jest.mock('../../components/ui/button', () => ({
  Button: ({ children, onClick, disabled, type, variant, size, asChild, className }: any) => {
    if (asChild) {
      return <div className={className}>{children}</div>;
    }
    return (
      <button 
        onClick={onClick} 
        disabled={disabled} 
        type={type}
        className={`button ${variant || ''} ${size || ''} ${className || ''}`}
      >
        {children}
      </button>
    );
  },
}));

jest.mock('../../components/ui/input', () => ({
  Input: ({ value, onChange, placeholder, required, disabled, id, className }: any) => (
    <input
      id={id}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      required={required}
      disabled={disabled}
      className={className}
    />
  ),
}));

jest.mock('../../components/ui/textarea', () => ({
  Textarea: ({ value, onChange, placeholder, disabled, id, rows, className }: any) => (
    <textarea
      id={id}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      disabled={disabled}
      rows={rows}
      className={className}
    />
  ),
}));

jest.mock('../../components/ui/badge', () => ({
  Badge: ({ children, className, style }: any) => (
    <span className={`badge ${className || ''}`} style={style}>
      {children}
    </span>
  ),
}));

// Mock Lucide React icons
jest.mock('lucide-react', () => ({
  ExternalLink: () => <span data-testid="external-link-icon">ExternalLink</span>,
  GitBranch: () => <span data-testid="git-branch-icon">GitBranch</span>,
  AlertCircle: () => <span data-testid="alert-circle-icon">AlertCircle</span>,
  GitPullRequest: () => <span data-testid="git-pull-request-icon">GitPullRequest</span>,
  Calendar: () => <span data-testid="calendar-icon">Calendar</span>,
  Users: () => <span data-testid="users-icon">Users</span>,
  Loader2: () => <span data-testid="loader-icon" className="animate-spin">Loader2</span>,
}));

const mockGitHubIssue: GitHubIssue = {
  id: 1,
  node_id: 'node_1',
  url: 'https://api.github.com/repos/test/repo/issues/1',
  repository_url: 'https://api.github.com/repos/test/repo',
  labels_url: 'https://api.github.com/repos/test/repo/issues/1/labels{/name}',
  comments_url: 'https://api.github.com/repos/test/repo/issues/1/comments',
  events_url: 'https://api.github.com/repos/test/repo/issues/1/events',
  html_url: 'https://github.com/test/repo/issues/1',
  number: 1,
  title: 'Test Issue',
  body: 'This is a test issue description',
  state: 'open',
  locked: false,
  user: {
    id: 1,
    login: 'testuser',
    node_id: 'user_1',
    avatar_url: 'https://github.com/images/error/testuser_happy.gif',
    gravatar_id: '',
    url: 'https://api.github.com/users/testuser',
    html_url: 'https://github.com/testuser',
    followers_url: 'https://api.github.com/users/testuser/followers',
    following_url: 'https://api.github.com/users/testuser/following{/other_user}',
    gists_url: 'https://api.github.com/users/testuser/gists{/gist_id}',
    starred_url: 'https://api.github.com/users/testuser/starred{/owner}{/repo}',
    subscriptions_url: 'https://api.github.com/users/testuser/subscriptions',
    organizations_url: 'https://api.github.com/users/testuser/orgs',
    repos_url: 'https://api.github.com/users/testuser/repos',
    events_url: 'https://api.github.com/users/testuser/events{/privacy}',
    received_events_url: 'https://api.github.com/users/testuser/received_events',
    type: 'User',
    site_admin: false,
  },
  labels: [{
    id: 1,
    node_id: 'label_1',
    url: 'https://api.github.com/repos/test/repo/labels/bug',
    name: 'bug',
    color: 'd73a49',
    default: true,
    description: 'Something isn\'t working',
  }],
  assignee: null,
  assignees: [],
  milestone: null,
  comments: 0,
  created_at: '2024-09-08T10:00:00Z',
  updated_at: '2024-09-08T10:00:00Z',
  closed_at: null,
  author_association: 'OWNER',
};

const mockGitHubPullRequest: GitHubPullRequest = {
  id: 2,
  node_id: 'pr_2',
  url: 'https://api.github.com/repos/test/repo/pulls/2',
  html_url: 'https://github.com/test/repo/pull/2',
  diff_url: 'https://github.com/test/repo/pull/2.diff',
  patch_url: 'https://github.com/test/repo/pull/2.patch',
  issue_url: 'https://api.github.com/repos/test/repo/issues/2',
  commits_url: 'https://api.github.com/repos/test/repo/pulls/2/commits',
  review_comments_url: 'https://api.github.com/repos/test/repo/pulls/2/comments',
  review_comment_url: 'https://api.github.com/repos/test/repo/pulls/comments{/number}',
  comments_url: 'https://api.github.com/repos/test/repo/issues/2/comments',
  statuses_url: 'https://api.github.com/repos/test/repo/statuses/abc123',
  number: 2,
  title: 'Test PR',
  body: 'This is a test pull request',
  state: 'open',
  locked: false,
  active_lock_reason: null,
  user: mockGitHubIssue.user,
  labels: [],
  milestone: null,
  assignee: null,
  assignees: [],
  requested_reviewers: [],
  requested_teams: [],
  head: {
    label: 'testuser:feature-branch',
    ref: 'feature-branch',
    sha: 'abc123',
    user: mockGitHubIssue.user,
    repo: {
      id: 1,
      node_id: 'repo_1',
      name: 'repo',
      full_name: 'test/repo',
      owner: mockGitHubIssue.user,
      private: false,
      html_url: 'https://github.com/test/repo',
      description: 'Test repository',
      fork: false,
      url: 'https://api.github.com/repos/test/repo',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-09-08T10:00:00Z',
      pushed_at: '2024-09-08T10:00:00Z',
      clone_url: 'https://github.com/test/repo.git',
      size: 1000,
      stargazers_count: 10,
      watchers_count: 10,
      language: 'TypeScript',
      forks_count: 5,
      archived: false,
      disabled: false,
      open_issues_count: 1,
      license: null,
      allow_forking: true,
      is_template: false,
      topics: [],
      visibility: 'public',
      forks: 5,
      open_issues: 1,
      watchers: 10,
      default_branch: 'main',
    },
  },
  base: {
    label: 'test:main',
    ref: 'main',
    sha: 'def456',
    user: mockGitHubIssue.user,
    repo: {
      id: 1,
      node_id: 'repo_1',
      name: 'repo',
      full_name: 'test/repo',
      owner: mockGitHubIssue.user,
      private: false,
      html_url: 'https://github.com/test/repo',
      description: 'Test repository',
      fork: false,
      url: 'https://api.github.com/repos/test/repo',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-09-08T10:00:00Z',
      pushed_at: '2024-09-08T10:00:00Z',
      clone_url: 'https://github.com/test/repo.git',
      size: 1000,
      stargazers_count: 10,
      watchers_count: 10,
      language: 'TypeScript',
      forks_count: 5,
      archived: false,
      disabled: false,
      open_issues_count: 1,
      license: null,
      allow_forking: true,
      is_template: false,
      topics: [],
      visibility: 'public',
      forks: 5,
      open_issues: 1,
      watchers: 10,
      default_branch: 'main',
    },
  },
  _links: {
    self: { href: 'https://api.github.com/repos/test/repo/pulls/2' },
    html: { href: 'https://github.com/test/repo/pull/2' },
    issue: { href: 'https://api.github.com/repos/test/repo/issues/2' },
    comments: { href: 'https://api.github.com/repos/test/repo/issues/2/comments' },
    review_comments: { href: 'https://api.github.com/repos/test/repo/pulls/2/comments' },
    review_comment: { href: 'https://api.github.com/repos/test/repo/pulls/comments{/number}' },
    commits: { href: 'https://api.github.com/repos/test/repo/pulls/2/commits' },
    statuses: { href: 'https://api.github.com/repos/test/repo/statuses/abc123' },
  },
  author_association: 'OWNER',
  auto_merge: null,
  draft: false,
  merged: false,
  mergeable: true,
  rebaseable: true,
  mergeable_state: 'clean',
  merged_by: null,
  comments: 0,
  review_comments: 0,
  maintainer_can_modify: true,
  commits: 1,
  additions: 10,
  deletions: 5,
  changed_files: 2,
  created_at: '2024-09-08T10:00:00Z',
  updated_at: '2024-09-08T10:00:00Z',
  closed_at: null,
  merged_at: null,
  merge_commit_sha: null,
};

describe('TaskCreationForm', () => {
  const mockOnSubmit = jest.fn();
  const mockOnCancel = jest.fn();

  beforeEach(() => {
    mockOnSubmit.mockClear();
    mockOnCancel.mockClear();
  });

  it('renders form for regular task creation', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
      />
    );

    expect(screen.getByLabelText(/task title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/branch name/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /create task/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('pre-populates form with GitHub issue data', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        githubIssue={mockGitHubIssue}
      />
    );

    const titleInput = screen.getByDisplayValue(/Issue #1: Test Issue/);
    const descriptionTextarea = screen.getByDisplayValue(/GitHub Issue: https:\/\/github\.com\/test\/repo\/issues\/1/);

    expect(titleInput).toBeInTheDocument();
    expect(descriptionTextarea).toBeInTheDocument();
    expect(screen.getByText('GitHub Issue #1')).toBeInTheDocument();
    expect(screen.getByText('bug')).toBeInTheDocument();
  });

  it('pre-populates form with GitHub PR data', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        githubPullRequest={mockGitHubPullRequest}
      />
    );

    const titleInput = screen.getByDisplayValue(/PR #2: Test PR/);
    const descriptionTextarea = screen.getByDisplayValue(/GitHub PR: https:\/\/github\.com\/test\/repo\/pull\/2/);
    
    // For PR, the branch name field is disabled and shows the PR's head branch
    const branchInput = screen.getByDisplayValue('feature-branch');

    expect(titleInput).toBeInTheDocument();
    expect(descriptionTextarea).toBeInTheDocument();
    expect(branchInput).toBeInTheDocument();
    expect(branchInput).toBeDisabled(); // Should be disabled for PR
    expect(screen.getByText('GitHub Pull Request #2')).toBeInTheDocument();
    expect(screen.getByText('feature-branch → main')).toBeInTheDocument();
  });

  it('calls onSubmit with correct data for regular task', async () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
      />
    );

    const titleInput = screen.getByLabelText(/task title/i);
    const descriptionTextarea = screen.getByLabelText(/description/i);
    const branchInput = screen.getByLabelText(/branch name/i);

    fireEvent.change(titleInput, { target: { value: 'New Task' } });
    fireEvent.change(descriptionTextarea, { target: { value: 'Task description' } });
    fireEvent.change(branchInput, { target: { value: 'feature/new-task' } });

    const submitButton = screen.getByRole('button', { name: /create task/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        title: 'New Task',
        description: 'Task description',
        status: 'pending',
        repository_id: 1,
        branch_name: 'feature/new-task',
        pr_url: undefined,
      });
    });
  });

  it('calls onSubmit with GitHub PR URL for PR-based task', async () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        githubPullRequest={mockGitHubPullRequest}
      />
    );

    const submitButton = screen.getByRole('button', { name: /create task/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        title: 'PR #2: Test PR',
        description: expect.stringContaining('GitHub PR: https://github.com/test/repo/pull/2'),
        status: 'pending',
        repository_id: 1,
        branch_name: 'feature-branch',
        pr_url: 'https://github.com/test/repo/pull/2',
      });
    });
  });

  it('calls onCancel when cancel button is clicked', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
      />
    );

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    fireEvent.click(cancelButton);

    expect(mockOnCancel).toHaveBeenCalledTimes(1);
  });

  it('disables form when isSubmitting is true', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isSubmitting={true}
      />
    );

    const titleInput = screen.getByLabelText(/task title/i);
    const descriptionTextarea = screen.getByLabelText(/description/i);
    const submitButton = screen.getByRole('button', { name: /create task/i });

    expect(titleInput).toBeDisabled();
    expect(descriptionTextarea).toBeDisabled();
    expect(submitButton).toBeDisabled();
    expect(screen.getByTestId('loader-icon')).toBeInTheDocument();
  });

  it('requires title field to be filled', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
      />
    );

    const submitButton = screen.getByRole('button', { name: /create task/i });
    expect(submitButton).toBeDisabled();

    const titleInput = screen.getByLabelText(/task title/i);
    fireEvent.change(titleInput, { target: { value: 'New Task' } });

    expect(submitButton).toBeEnabled();
  });

  it('displays GitHub metadata correctly for issues', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        githubIssue={mockGitHubIssue}
      />
    );

    expect(screen.getByText('GitHub Issue #1')).toBeInTheDocument();
    expect(screen.getByText('open')).toBeInTheDocument();
    expect(screen.getByText('bug')).toBeInTheDocument();
    expect(screen.getByText('by testuser')).toBeInTheDocument();
  });

  it('displays GitHub metadata correctly for PRs', () => {
    render(
      <TaskCreationForm
        repositoryId={1}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        githubPullRequest={mockGitHubPullRequest}
      />
    );

    expect(screen.getByText('GitHub Pull Request #2')).toBeInTheDocument();
    expect(screen.getByText('open')).toBeInTheDocument();
    expect(screen.getByText('feature-branch → main')).toBeInTheDocument();
    expect(screen.getByText('by testuser')).toBeInTheDocument();
  });
});