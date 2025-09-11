/**
 * @jest-environment jsdom
 */

import { WorkLogManager } from '../../services/WorkLogManager';

// Mock fetch globally
global.fetch = jest.fn();

describe('WorkLogManager', () => {
  let workLogManager: WorkLogManager;
  let mockFetch: jest.MockedFunction<typeof fetch>;

  beforeEach(() => {
    workLogManager = WorkLogManager.getInstance();
    mockFetch = fetch as jest.MockedFunction<typeof fetch>;
    mockFetch.mockClear();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe('Task Creation Logging', () => {
    it('should log task creation successfully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as Response);

      await workLogManager.logTaskCreated(
        'task-123',
        'Test Task',
        'test-repo',
        {
          branch: 'feature/test',
          githubIssue: 42,
          description: 'Test task description'
        }
      );

      expect(mockFetch).toHaveBeenCalledTimes(1);
      expect(mockFetch).toHaveBeenCalledWith('/api/work-logs/entry', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: expect.stringContaining('"taskId":"task-123"')
      });

      const callBody = JSON.parse(mockFetch.mock.calls[0][1]?.body as string);
      expect(callBody.repository).toBe('test-repo');
      expect(callBody.entry.taskTitle).toBe('Test Task');
      expect(callBody.entry.status).toBe('pending');
      expect(callBody.entry.metadata.branch).toBe('feature/test');
      expect(callBody.entry.metadata.githubIssue).toBe(42);
    });

    it('should handle logging errors gracefully', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      // Should not throw
      await expect(workLogManager.logTaskCreated(
        'task-123',
        'Test Task', 
        'test-repo'
      )).resolves.toBeUndefined();
    });
  });

  describe('Task Status Changes', () => {
    it('should log status changes with progress updates', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as Response);

      await workLogManager.logTaskStatusChange(
        'task-123',
        'Test Task',
        'test-repo',
        'in_progress',
        'Started working on implementation',
        ['Issue with API authentication'],
        ['Added better error handling']
      );

      expect(mockFetch).toHaveBeenCalledTimes(1);
      
      const callBody = JSON.parse(mockFetch.mock.calls[0][1]?.body as string);
      expect(callBody.entry.status).toBe('in_progress');
      expect(callBody.entry.progressUpdate).toBe('Started working on implementation');
      expect(callBody.entry.issuesDiscovered).toEqual(['Issue with API authentication']);
      expect(callBody.entry.improvementsMade).toEqual(['Added better error handling']);
    });
  });

  describe('Progress Logging', () => {
    it('should log progress updates with metadata', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as Response);

      await workLogManager.logProgress(
        'task-123',
        'Test Task',
        'test-repo',
        'Completed initial implementation',
        {
          tokensUsed: 1500,
          prUrl: 'https://github.com/test/repo/pull/1'
        }
      );

      const callBody = JSON.parse(mockFetch.mock.calls[0][1]?.body as string);
      expect(callBody.entry.progressUpdate).toBe('Completed initial implementation');
      expect(callBody.entry.metadata.tokensUsed).toBe(1500);
      expect(callBody.entry.metadata.prUrl).toBe('https://github.com/test/repo/pull/1');
    });
  });

  describe('Get Work Logs', () => {
    it('should fetch work logs for repository', async () => {
      const mockLogs = [
        {
          date: '2025-09-10',
          repository: 'test-repo',
          entries: []
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockLogs,
      } as Response);

      const result = await workLogManager.getWorkLogs('test-repo');

      expect(mockFetch).toHaveBeenCalledWith('/api/work-logs?repository=test-repo');
      expect(result).toEqual(mockLogs);
    });

    it('should handle fetch errors by returning empty array', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      const result = await workLogManager.getWorkLogs('test-repo');
      expect(result).toEqual([]);
    });
  });

  describe('Singleton Pattern', () => {
    it('should return the same instance', () => {
      const instance1 = WorkLogManager.getInstance();
      const instance2 = WorkLogManager.getInstance();
      
      expect(instance1).toBe(instance2);
    });
  });
});