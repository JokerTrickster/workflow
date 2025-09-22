import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { jest } from '@jest/globals';
import '@testing-library/jest-dom';

import TaskHistory from '../../components/TaskHistory/TaskHistory';
import { claudeService, WorkflowHistory, TaskHistoryResponse } from '../../services/claudeService';

// Mock the claude service
jest.mock('../../services/claudeService');
const mockClaudeService = claudeService as jest.Mocked<typeof claudeService>;

describe('TaskHistory Integration Tests', () => {
  const mockTaskHistoryData: WorkflowHistory[] = [
    {
      id: 1,
      request_id: 'req-1',
      status: 'completed',
      tasks: 'Implement user authentication',
      repository_name: 'test-repo',
      working_dir: '/path/to/repo',
      claude_cmd: 'claude code',
      interactive: true,
      continue_task: false,
      created_at: '2025-09-22T10:00:00Z',
      completed_at: '2025-09-22T10:30:00Z',
      processing_time_ms: 1800000,
      result: 'Authentication system implemented successfully'
    },
    {
      id: 2,
      request_id: 'req-2',
      status: 'processing',
      tasks: 'Add database migration',
      repository_name: 'test-repo',
      interactive: false,
      continue_task: false,
      created_at: '2025-09-22T11:00:00Z'
    },
    {
      id: 3,
      request_id: 'req-3',
      status: 'failed',
      tasks: 'Fix security vulnerability',
      repository_name: 'test-repo',
      interactive: true,
      continue_task: false,
      created_at: '2025-09-22T12:00:00Z',
      error: 'Security patch failed: dependencies conflict'
    },
    {
      id: 4,
      request_id: 'req-4',
      status: 'pending',
      tasks: 'Update documentation',
      repository_name: 'test-repo',
      interactive: false,
      continue_task: true,
      created_at: '2025-09-22T13:00:00Z'
    }
  ];

  const mockPaginationResponse: TaskHistoryResponse = {
    data: mockTaskHistoryData.slice(0, 3), // First page
    pagination: {
      page: 1,
      limit: 20,
      total: 25,
      total_pages: 2
    }
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockClaudeService.getTaskHistory.mockResolvedValue(mockPaginationResponse);
  });

  describe('End-to-End Integration', () => {
    test('loads and displays task history with all components working together', async () => {
      render(<TaskHistory repositoryName="test-repo" />);

      // Verify loading state appears first
      expect(screen.getByText(/loading task history/i)).toBeInTheDocument();

      // Wait for data to load
      await waitFor(() => {
        expect(screen.queryByText(/loading task history/i)).not.toBeInTheDocument();
      });

      // Verify API was called correctly
      expect(mockClaudeService.getTaskHistory).toHaveBeenCalledWith('test-repo', { page: 1, limit: 20 });

      // Verify all task items are rendered with correct status badges
      expect(screen.getByText('Implement user authentication')).toBeInTheDocument();
      expect(screen.getByText('Add database migration')).toBeInTheDocument();
      expect(screen.getByText('Fix security vulnerability')).toBeInTheDocument();

      // Verify status badges are rendered correctly
      expect(screen.getByText('Completed')).toBeInTheDocument();
      expect(screen.getByText('Processing')).toBeInTheDocument();
      expect(screen.getByText('Failed')).toBeInTheDocument();

      // Verify pagination is rendered
      expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      expect(screen.getByText('Total: 25 tasks')).toBeInTheDocument();
    });

    test('handles pagination navigation correctly', async () => {
      const mockPage2Response: TaskHistoryResponse = {
        data: [mockTaskHistoryData[3]], // Second page
        pagination: {
          page: 2,
          limit: 20,
          total: 25,
          total_pages: 2
        }
      };

      mockClaudeService.getTaskHistory
        .mockResolvedValueOnce(mockPaginationResponse) // First page
        .mockResolvedValueOnce(mockPage2Response); // Second page

      render(<TaskHistory repositoryName="test-repo" />);

      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByText('Implement user authentication')).toBeInTheDocument();
      });

      // Click next page button
      const nextButton = screen.getByRole('button', { name: /next/i });
      fireEvent.click(nextButton);

      // Wait for page 2 data to load
      await waitFor(() => {
        expect(screen.getByText('Update documentation')).toBeInTheDocument();
      });

      // Verify API was called for page 2
      expect(mockClaudeService.getTaskHistory).toHaveBeenCalledWith('test-repo', { page: 2, limit: 20 });

      // Verify pagination updated
      expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
    });

    test('displays detailed task information correctly', async () => {
      render(<TaskHistory repositoryName="test-repo" />);

      await waitFor(() => {
        expect(screen.getByText('Implement user authentication')).toBeInTheDocument();
      });

      // Verify completed task shows all details
      const completedTask = screen.getByText('Implement user authentication').closest('[data-testid="task-item"]');
      expect(completedTask).toBeInTheDocument();

      if (completedTask) {
        expect(completedTask).toHaveTextContent('req-1');
        expect(completedTask).toHaveTextContent('Completed');
        expect(completedTask).toHaveTextContent('30m 0s'); // Processing time
        expect(completedTask).toHaveTextContent('Authentication system implemented successfully');
        expect(completedTask).toHaveTextContent('Interactive'); // Interactive flag
      }

      // Verify failed task shows error
      const failedTask = screen.getByText('Fix security vulnerability').closest('[data-testid="task-item"]');
      if (failedTask) {
        expect(failedTask).toHaveTextContent('Security patch failed: dependencies conflict');
      }
    });

    test('handles real-time polling for processing tasks', async () => {
      // Start with a processing task
      const processingResponse: TaskHistoryResponse = {
        data: [{
          ...mockTaskHistoryData[1],
          status: 'processing'
        }],
        pagination: {
          page: 1,
          limit: 20,
          total: 1,
          total_pages: 1
        }
      };

      // Then simulate completion
      const completedResponse: TaskHistoryResponse = {
        data: [{
          ...mockTaskHistoryData[1],
          status: 'completed',
          completed_at: '2025-09-22T11:05:00Z',
          processing_time_ms: 300000,
          result: 'Database migration completed successfully'
        }],
        pagination: {
          page: 1,
          limit: 20,
          total: 1,
          total_pages: 1
        }
      };

      mockClaudeService.getTaskHistory
        .mockResolvedValueOnce(processingResponse)
        .mockResolvedValueOnce(completedResponse);

      render(<TaskHistory repositoryName="test-repo" pollingEnabled={true} pollingInterval={1000} />);

      // Wait for initial processing state
      await waitFor(() => {
        expect(screen.getByText('Processing')).toBeInTheDocument();
      });

      // Wait for polling to update to completed state
      await waitFor(() => {
        expect(screen.getByText('Completed')).toBeInTheDocument();
      }, { timeout: 2000 });

      // Verify the task updated with result
      expect(screen.getByText('Database migration completed successfully')).toBeInTheDocument();
    });
  });

  describe('Error Handling Integration', () => {
    test('handles API errors gracefully', async () => {
      const mockError = new Error('Network error: Failed to fetch task history');
      mockClaudeService.getTaskHistory.mockRejectedValue(mockError);

      render(<TaskHistory repositoryName="test-repo" />);

      // Wait for error to be displayed
      await waitFor(() => {
        expect(screen.getByText(/failed to load task history/i)).toBeInTheDocument();
      });

      expect(screen.getByText(/network error/i)).toBeInTheDocument();
    });

    test('handles empty repository gracefully', async () => {
      const emptyResponse: TaskHistoryResponse = {
        data: [],
        pagination: {
          page: 1,
          limit: 20,
          total: 0,
          total_pages: 0
        }
      };

      mockClaudeService.getTaskHistory.mockResolvedValue(emptyResponse);

      render(<TaskHistory repositoryName="empty-repo" />);

      await waitFor(() => {
        expect(screen.getByText(/no tasks found/i)).toBeInTheDocument();
      });

      expect(screen.getByText(/no task history available for this repository/i)).toBeInTheDocument();
    });

    test('handles network timeout errors', async () => {
      const timeoutError = new Error('Request timeout');
      timeoutError.name = 'TimeoutError';
      mockClaudeService.getTaskHistory.mockRejectedValue(timeoutError);

      render(<TaskHistory repositoryName="test-repo" />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load task history/i)).toBeInTheDocument();
      });

      // Should provide retry option
      const retryButton = screen.getByRole('button', { name: /retry/i });
      expect(retryButton).toBeInTheDocument();

      // Test retry functionality
      mockClaudeService.getTaskHistory.mockResolvedValueOnce(mockPaginationResponse);
      fireEvent.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText('Implement user authentication')).toBeInTheDocument();
      });
    });
  });

  describe('Performance Integration', () => {
    test('handles large datasets efficiently', async () => {
      // Create large dataset
      const largeDataset: WorkflowHistory[] = Array.from({ length: 100 }, (_, i) => ({
        id: i + 1,
        request_id: `large-req-${i}`,
        status: i % 4 === 0 ? 'completed' : i % 4 === 1 ? 'processing' : i % 4 === 2 ? 'failed' : 'pending',
        tasks: `Large dataset task ${i}`,
        repository_name: 'large-repo',
        interactive: i % 2 === 0,
        continue_task: false,
        created_at: new Date(Date.now() - i * 60000).toISOString()
      }));

      const largeResponse: TaskHistoryResponse = {
        data: largeDataset.slice(0, 20), // First page
        pagination: {
          page: 1,
          limit: 20,
          total: 100,
          total_pages: 5
        }
      };

      mockClaudeService.getTaskHistory.mockResolvedValue(largeResponse);

      const renderStart = performance.now();
      render(<TaskHistory repositoryName="large-repo" />);

      await waitFor(() => {
        expect(screen.getByText('Large dataset task 0')).toBeInTheDocument();
      });

      const renderEnd = performance.now();
      const renderTime = renderEnd - renderStart;

      // Should render within reasonable time (less than 1 second)
      expect(renderTime).toBeLessThan(1000);

      // Should display pagination for large dataset
      expect(screen.getByText('Page 1 of 5')).toBeInTheDocument();
      expect(screen.getByText('Total: 100 tasks')).toBeInTheDocument();
    });

    test('debounces rapid pagination clicks', async () => {
      render(<TaskHistory repositoryName="test-repo" />);

      await waitFor(() => {
        expect(screen.getByText('Implement user authentication')).toBeInTheDocument();
      });

      const nextButton = screen.getByRole('button', { name: /next/i });

      // Rapidly click next button multiple times
      fireEvent.click(nextButton);
      fireEvent.click(nextButton);
      fireEvent.click(nextButton);

      // Should only make one additional API call despite multiple clicks
      await waitFor(() => {
        expect(mockClaudeService.getTaskHistory).toHaveBeenCalledTimes(2); // Initial + 1 debounced call
      });
    });
  });

  describe('Concurrent Operations Integration', () => {
    test('handles concurrent component instances correctly', async () => {
      const repo1Response: TaskHistoryResponse = {
        data: [mockTaskHistoryData[0]],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 }
      };

      const repo2Response: TaskHistoryResponse = {
        data: [mockTaskHistoryData[1]],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 }
      };

      mockClaudeService.getTaskHistory
        .mockImplementation((repoName: string) => {
          if (repoName === 'repo1') return Promise.resolve(repo1Response);
          if (repoName === 'repo2') return Promise.resolve(repo2Response);
          return Promise.reject(new Error('Unknown repository'));
        });

      const { container } = render(
        <div>
          <TaskHistory repositoryName="repo1" />
          <TaskHistory repositoryName="repo2" />
        </div>
      );

      // Wait for both components to load
      await waitFor(() => {
        expect(screen.getByText('Implement user authentication')).toBeInTheDocument();
        expect(screen.getByText('Add database migration')).toBeInTheDocument();
      });

      // Verify both API calls were made with correct parameters
      expect(mockClaudeService.getTaskHistory).toHaveBeenCalledWith('repo1', { page: 1, limit: 20 });
      expect(mockClaudeService.getTaskHistory).toHaveBeenCalledWith('repo2', { page: 1, limit: 20 });

      // Verify components are isolated (each shows its own data)
      const taskHistoryComponents = container.querySelectorAll('[data-testid="task-history"]');
      expect(taskHistoryComponents).toHaveLength(2);
    });

    test('handles repository switching without memory leaks', async () => {
      const { rerender } = render(<TaskHistory repositoryName="repo1" />);

      await waitFor(() => {
        expect(mockClaudeService.getTaskHistory).toHaveBeenCalledWith('repo1', { page: 1, limit: 20 });
      });

      // Switch repository
      rerender(<TaskHistory repositoryName="repo2" />);

      await waitFor(() => {
        expect(mockClaudeService.getTaskHistory).toHaveBeenCalledWith('repo2', { page: 1, limit: 20 });
      });

      // Should not continue polling for old repository
      expect(mockClaudeService.getTaskHistory).toHaveBeenCalledTimes(2);
    });
  });

  describe('Status Update Integration', () => {
    test('reflects real-time status changes correctly', async () => {
      const initialResponse: TaskHistoryResponse = {
        data: [{
          ...mockTaskHistoryData[1],
          status: 'pending'
        }],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 }
      };

      const processingResponse: TaskHistoryResponse = {
        data: [{
          ...mockTaskHistoryData[1],
          status: 'processing'
        }],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 }
      };

      const completedResponse: TaskHistoryResponse = {
        data: [{
          ...mockTaskHistoryData[1],
          status: 'completed',
          completed_at: '2025-09-22T11:10:00Z',
          processing_time_ms: 600000,
          result: 'Database migration completed'
        }],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 }
      };

      mockClaudeService.getTaskHistory
        .mockResolvedValueOnce(initialResponse)
        .mockResolvedValueOnce(processingResponse)
        .mockResolvedValueOnce(completedResponse);

      render(<TaskHistory repositoryName="test-repo" pollingEnabled={true} pollingInterval={500} />);

      // Initial state: pending
      await waitFor(() => {
        expect(screen.getByText('Pending')).toBeInTheDocument();
      });

      // After first poll: processing
      await waitFor(() => {
        expect(screen.getByText('Processing')).toBeInTheDocument();
      }, { timeout: 1000 });

      // After second poll: completed
      await waitFor(() => {
        expect(screen.getByText('Completed')).toBeInTheDocument();
        expect(screen.getByText('Database migration completed')).toBeInTheDocument();
      }, { timeout: 1000 });
    });
  });
});