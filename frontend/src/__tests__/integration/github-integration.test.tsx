/**
 * Comprehensive integration test for GitHub integration workflow
 * Tests the complete flow: connect repository → view tabs → create task → view logs/dashboard
 */

import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient } from '@tanstack/react-query';
import { QueryProvider } from '../../providers/QueryProvider';
import { WorkspacePanel } from '../../presentation/components/WorkspacePanel';
import { GitHubApiService } from '../../services/githubApi';
import { ActivityLogger } from '../../services/ActivityLogger';
import { ErrorHandler, ErrorType } from '../../utils/errorHandler';
import { Repository } from '../../domain/entities/Repository';
import '@testing-library/jest-dom';

// Mock external dependencies
jest.mock('../../services/githubApi');
jest.mock('../../lib/supabase/client', () => ({
  supabase: {
    auth: {
      getSession: jest.fn(),
      onAuthStateChange: jest.fn(() => ({ data: { subscription: { unsubscribe: jest.fn() } } }))
    }
  }
}));

const mockGitHubApiService = GitHubApiService as jest.Mocked<typeof GitHubApiService>;

// Mock ActivityLogger
const mockActivityLoggerInstance = {
  logTaskCreated: jest.fn(),
  logTaskStarted: jest.fn(),
  logTaskFailed: jest.fn(),
  logGitHubSync: jest.fn(),
  logGitHubApiCall: jest.fn(),
  logGitHubRateLimit: jest.fn(),
};

jest.mock('../../services/ActivityLogger', () => ({
  ActivityLogger: {
    getInstance: jest.fn(() => mockActivityLoggerInstance),
  }
}));

// Test data
const mockRepository: Repository = {
  id: 123,
  name: 'test-repo',
  full_name: 'testowner/test-repo',
  description: 'A test repository for integration testing',
  html_url: 'https://github.com/testowner/test-repo',
  clone_url: 'https://github.com/testowner/test-repo.git',
  ssh_url: 'git@github.com:testowner/test-repo.git',
  private: false,
  language: 'TypeScript',
  stargazers_count: 42,
  forks_count: 7,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-09-09T00:00:00Z',
  pushed_at: '2024-09-09T00:00:00Z',
  default_branch: 'main',
  is_connected: true,
};

const mockIssues = [
  {
    id: 1,
    number: 1,
    title: 'Add new feature',
    body: 'This is a test issue for adding a new feature',
    state: 'open',
    state_reason: null,
    user: { login: 'testuser', avatar_url: 'https://github.com/testuser.png', id: 1 },
    assignees: [],
    assignee: null,
    labels: [{ id: 1, name: 'enhancement', color: '0075ca' }],
    comments: 2,
    html_url: 'https://github.com/testowner/test-repo/issues/1',
    created_at: '2024-09-01T00:00:00Z',
    updated_at: '2024-09-01T00:00:00Z',
  },
];

const mockPullRequests = [
  {
    id: 2,
    number: 2,
    title: 'Fix authentication bug',
    body: 'This PR fixes the authentication issue',
    state: 'open',
    merged: false,
    draft: false,
    user: { login: 'testuser', avatar_url: 'https://github.com/testuser.png', id: 1 },
    assignees: [],
    assignee: null,
    requested_reviewers: [],
    labels: [{ id: 2, name: 'bug', color: 'd73a49' }],
    comments: 1,
    html_url: 'https://github.com/testowner/test-repo/pull/2',
    head: { ref: 'fix-auth', sha: 'abc123' },
    base: { ref: 'main', sha: 'def456' },
    additions: 15,
    deletions: 8,
    commits: 3,
    changed_files: 2,
    created_at: '2024-09-02T00:00:00Z',
    updated_at: '2024-09-02T00:00:00Z',
  },
];

// Custom render function with providers - simplified to avoid QueryClient issues
const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <QueryProvider>
      {ui}
    </QueryProvider>
  );
};

// Utility to simulate network errors
const simulateNetworkError = () => {
  return Promise.reject(new Error('Network error'));
};

// Utility to simulate rate limit error
const simulateRateLimitError = () => {
  const error = new Error('GitHub API rate limit exceeded. Resets at 10:00 AM.');
  (error as any).status = 403;
  return Promise.reject(error);
};

describe('GitHub Integration Workflow', () => {
  let user: ReturnType<typeof userEvent.setup>;

  beforeEach(() => {
    user = userEvent.setup();
    
    // Reset all mocks
    jest.clearAllMocks();
    
    // Default successful API responses
    mockGitHubApiService.fetchRepositoryIssues.mockResolvedValue({
      issues: mockIssues,
      nextCursor: undefined,
      hasMore: false,
      pagination: {
        page: 1,
        per_page: 30,
        has_next_page: false,
        has_previous_page: false,
      },
    });

    mockGitHubApiService.fetchRepositoryPullRequests.mockResolvedValue({
      pullRequests: mockPullRequests,
      nextCursor: undefined,
      hasMore: false,
      pagination: {
        page: 1,
        per_page: 30,
        has_next_page: false,
        has_previous_page: false,
      },
    });
  });

  describe('Happy Path: Complete Workflow', () => {
    it('should complete the full workflow: view repository → navigate tabs → create task → view logs/dashboard', async () => {
      const onClose = jest.fn();
      
      renderWithProviders(
        <WorkspacePanel repository={mockRepository} onClose={onClose} />
      );

      // Step 1: Verify connected repository display
      expect(screen.getByText('test-repo')).toBeInTheDocument();
      expect(screen.getByText('Connected')).toBeInTheDocument();
      expect(screen.getByText('A test repository for integration testing')).toBeInTheDocument();

      // Step 2: Navigate to Issues tab and view GitHub issues
      const issuesTab = screen.getByRole('tab', { name: /Issues/i });
      await user.click(issuesTab);

      await waitFor(() => {
        expect(screen.getByText('GitHub Issues')).toBeInTheDocument();
      });

      // Wait for issues to load
      await waitFor(() => {
        expect(screen.getByText('#1: Add new feature')).toBeInTheDocument();
      });

      // Step 3: Create task from GitHub issue
      const createTaskButton = screen.getByRole('button', { name: /Create Task/i });
      await user.click(createTaskButton);

      await waitFor(() => {
        expect(screen.getByText('Create Task from GitHub')).toBeInTheDocument();
      });

      // Fill in task creation form
      const titleInput = screen.getByDisplayValue('Add new feature');
      expect(titleInput).toBeInTheDocument();

      const descriptionTextarea = screen.getByDisplayValue('This is a test issue for adding a new feature');
      expect(descriptionTextarea).toBeInTheDocument();

      // Submit task creation
      const submitButton = screen.getByRole('button', { name: /Create Task/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockActivityLoggerInstance.logTaskCreated).toHaveBeenCalled();
      });

      // Step 4: Verify task appears in Tasks tab
      const tasksTab = screen.getByRole('tab', { name: /Tasks/i });
      await user.click(tasksTab);

      await waitFor(() => {
        expect(screen.getByText('Add new feature')).toBeInTheDocument();
      });

      // Step 5: Navigate to Logs tab
      const logsTab = screen.getByRole('tab', { name: /Logs/i });
      await user.click(logsTab);

      await waitFor(() => {
        expect(screen.getByText(/Activity Logs/i)).toBeInTheDocument();
      });

      // Step 6: Navigate to Dashboard tab
      const dashboardTab = screen.getByRole('tab', { name: /Dashboard/i });
      await user.click(dashboardTab);

      await waitFor(() => {
        expect(screen.getByText(/Dashboard/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling: GitHub API Failures', () => {
    it('should handle GitHub API network errors gracefully', async () => {
      // Simulate network error for issues
      mockGitHubApiService.fetchRepositoryIssues.mockRejectedValue(
        new Error('Failed to fetch data from GitHub API')
      );

      const onClose = jest.fn();
      
      renderWithProviders(
        <WorkspacePanel repository={mockRepository} onClose={onClose} />
      );

      // Navigate to Issues tab
      const issuesTab = screen.getByRole('tab', { name: /Issues/i });
      await user.click(issuesTab);

      // Should show error state
      await waitFor(() => {
        expect(screen.getByText('Failed to load issues')).toBeInTheDocument();
        expect(screen.getByText('Failed to fetch data from GitHub API')).toBeInTheDocument();
      });

      // Should show retry button
      const retryButton = screen.getByRole('button', { name: /Try Again/i });
      expect(retryButton).toBeInTheDocument();

      // Test retry functionality
      mockGitHubApiService.fetchRepositoryIssues.mockResolvedValue({
        issues: mockIssues,
        nextCursor: undefined,
        hasMore: false,
        pagination: {
          page: 1,
          per_page: 30,
          has_next_page: false,
          has_previous_page: false,
        },
      });

      await user.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText('#1: Add new feature')).toBeInTheDocument();
      });
    });

    it('should handle GitHub API rate limiting with user-friendly messages', async () => {
      // Simulate rate limit error
      const rateLimitError = new Error('GitHub API rate limit exceeded. Resets at 10:00 AM.');
      (rateLimitError as any).status = 403;
      mockGitHubApiService.fetchRepositoryIssues.mockRejectedValue(rateLimitError);

      const onClose = jest.fn();
      
      renderWithProviders(
        <WorkspacePanel repository={mockRepository} onClose={onClose} />
      );

      // Navigate to Issues tab
      const issuesTab = screen.getByRole('tab', { name: /Issues/i });
      await user.click(issuesTab);

      // Should show rate limit error message
      await waitFor(() => {
        expect(screen.getByText('Failed to load issues')).toBeInTheDocument();
        expect(screen.getByText(/rate limit exceeded/i)).toBeInTheDocument();
      });
    });
  });

  describe('Graceful Degradation', () => {
    it('should show appropriate message for unconnected repositories', async () => {
      const unconnectedRepo: Repository = {
        ...mockRepository,
        is_connected: false,
      };

      const onClose = jest.fn();
      
      renderWithProviders(
        <WorkspacePanel repository={unconnectedRepo} onClose={onClose} />
      );

      // Should show not connected message
      expect(screen.getByText('Repository Not Connected')).toBeInTheDocument();
      expect(screen.getByText(/needs to be connected before you can access/i)).toBeInTheDocument();
      
      // Should show action buttons
      expect(screen.getByRole('button', { name: /Back to Repositories/i })).toBeInTheDocument();
      expect(screen.getByRole('link', { name: /View on GitHub/i })).toBeInTheDocument();
    });

    it('should handle empty data states appropriately', async () => {
      // Mock empty responses
      mockGitHubApiService.fetchRepositoryIssues.mockResolvedValue({
        issues: [],
        nextCursor: undefined,
        hasMore: false,
        pagination: {
          page: 1,
          per_page: 30,
          has_next_page: false,
          has_previous_page: false,
        },
      });

      const onClose = jest.fn();
      
      renderWithProviders(
        <WorkspacePanel repository={mockRepository} onClose={onClose} />
      );

      const issuesTab = screen.getByRole('tab', { name: /Issues/i });
      await user.click(issuesTab);

      await waitFor(() => {
        expect(screen.getByText('No issues found')).toBeInTheDocument();
        expect(screen.getByText(/This repository has no open issues/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Recovery', () => {
    it('should allow users to retry failed operations', async () => {
      // First call fails, second succeeds
      mockGitHubApiService.fetchRepositoryIssues
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValue({
          issues: mockIssues,
          nextCursor: undefined,
          hasMore: false,
          pagination: {
            page: 1,
            per_page: 30,
            has_next_page: false,
            has_previous_page: false,
          },
        });

      const onClose = jest.fn();
      
      renderWithProviders(
        <WorkspacePanel repository={mockRepository} onClose={onClose} />
      );

      const issuesTab = screen.getByRole('tab', { name: /Issues/i });
      await user.click(issuesTab);

      // Should show error
      await waitFor(() => {
        expect(screen.getByText('Failed to load issues')).toBeInTheDocument();
      });

      // Click retry
      const retryButton = screen.getByRole('button', { name: /Try Again/i });
      await user.click(retryButton);

      // Should now show data
      await waitFor(() => {
        expect(screen.getByText('#1: Add new feature')).toBeInTheDocument();
      });

      // Verify the API was called twice
      expect(mockGitHubApiService.fetchRepositoryIssues).toHaveBeenCalledTimes(2);
    });
  });
});