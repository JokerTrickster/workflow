/**
 * Basic tests for ActivityLogger functionality
 * This ensures the activity logging system works correctly
 */

import { ActivityLogger } from '../../services/ActivityLogger';
import { ActivityType, ActivityLevel } from '../../types/activity';

// Mock localStorage for tests
const localStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    }
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

describe('ActivityLogger', () => {
  let logger: ActivityLogger;

  beforeEach(() => {
    localStorageMock.clear();
    // Reset singleton instance
    (ActivityLogger as any).instance = null;
    logger = ActivityLogger.getInstance();
  });

  describe('Singleton Pattern', () => {
    it('should return the same instance', () => {
      const logger1 = ActivityLogger.getInstance();
      const logger2 = ActivityLogger.getInstance();
      expect(logger1).toBe(logger2);
    });
  });

  describe('Basic Logging', () => {
    it('should log activities correctly', () => {
      logger.log('task', 'info', 'Test Task', 'Test description');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].title).toBe('Test Task');
      expect(logs[0].description).toBe('Test description');
      expect(logs[0].type).toBe('task');
      expect(logs[0].level).toBe('info');
    });

    it('should generate unique IDs for each log', () => {
      logger.log('task', 'info', 'Task 1', 'Description 1');
      logger.log('task', 'success', 'Task 2', 'Description 2');
      
      const logs = logger.getLogs();
      expect(logs[0].id).not.toBe(logs[1].id);
    });

    it('should maintain chronological order (newest first)', () => {
      logger.log('task', 'info', 'First Task', 'First');
      logger.log('task', 'info', 'Second Task', 'Second');
      
      const logs = logger.getLogs();
      expect(logs[0].title).toBe('Second Task');
      expect(logs[1].title).toBe('First Task');
    });
  });

  describe('Repository Events', () => {
    it('should log repository connection', () => {
      logger.logRepositoryConnected(123, 'test-repo', '/path/to/repo');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('connection');
      expect(logs[0].level).toBe('success');
      expect(logs[0].title).toBe('Repository connected');
      expect(logs[0].metadata?.repositoryId).toBe(123);
      expect(logs[0].metadata?.repositoryName).toBe('test-repo');
    });

    it('should log repository disconnection', () => {
      logger.logRepositoryDisconnected(123, 'test-repo');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('connection');
      expect(logs[0].level).toBe('info');
      expect(logs[0].title).toBe('Repository disconnected');
    });

    it('should log repository connection failure', () => {
      logger.logRepositoryConnectionFailed('test-repo', 'Network error');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('connection');
      expect(logs[0].level).toBe('error');
      expect(logs[0].metadata?.errorMessage).toBe('Network error');
    });
  });

  describe('Task Events', () => {
    it('should log task creation', () => {
      logger.logTaskCreated('task-1', 'Test Task', 123, 'test-repo');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('task');
      expect(logs[0].level).toBe('info');
      expect(logs[0].title).toBe('Task created');
      expect(logs[0].metadata?.taskId).toBe('task-1');
      expect(logs[0].metadata?.taskTitle).toBe('Test Task');
    });

    it('should log task start', () => {
      logger.logTaskStarted('task-1', 'Test Task');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('task');
      expect(logs[0].level).toBe('info');
      expect(logs[0].title).toBe('Task started');
    });

    it('should log task completion', () => {
      logger.logTaskCompleted('task-1', 'Test Task', 5000);
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('task');
      expect(logs[0].level).toBe('success');
      expect(logs[0].title).toBe('Task completed');
      expect(logs[0].metadata?.duration).toBe(5000);
    });

    it('should log task failure', () => {
      logger.logTaskFailed('task-1', 'Test Task', 'Build failed');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('task');
      expect(logs[0].level).toBe('error');
      expect(logs[0].title).toBe('Task failed');
      expect(logs[0].metadata?.errorMessage).toBe('Build failed');
    });
  });

  describe('GitHub Events', () => {
    it('should log GitHub sync events', () => {
      logger.logGitHubSync('test-repo', 'started');
      logger.logGitHubSync('test-repo', 'completed', { duration: 1000, apiCallCount: 3 });
      logger.logGitHubSync('test-repo', 'failed', { error: 'Network timeout' });
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(3);
      
      expect(logs[2].level).toBe('info');
      expect(logs[1].level).toBe('success');
      expect(logs[0].level).toBe('error');
    });

    it('should log GitHub API calls', () => {
      logger.logGitHubApiCall('/repos/owner/repo', 'GET', 4000);
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('github');
      expect(logs[0].metadata?.rateLimitRemaining).toBe(4000);
    });

    it('should log rate limit warnings', () => {
      logger.logGitHubRateLimit(50, '14:30');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('github');
      expect(logs[0].level).toBe('warning');
      expect(logs[0].metadata?.rateLimitRemaining).toBe(50);
    });
  });

  describe('Navigation Events', () => {
    it('should log tab switches', () => {
      logger.logTabSwitch('tasks', 'issues', 'test-repo');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('navigation');
      expect(logs[0].metadata?.previousTab).toBe('tasks');
      expect(logs[0].metadata?.currentTab).toBe('issues');
    });

    it('should log workspace access', () => {
      logger.logWorkspaceAccess('test-repo', 'opened');
      
      const logs = logger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].type).toBe('navigation');
      expect(logs[0].title).toBe('Workspace opened');
    });
  });

  describe('Filtering', () => {
    beforeEach(() => {
      logger.log('connection', 'success', 'Connection Test', 'Connected');
      logger.log('task', 'info', 'Task Test', 'Created task');
      logger.log('github', 'warning', 'GitHub Test', 'Rate limit warning');
      logger.log('navigation', 'info', 'Navigation Test', 'Tab switched');
    });

    it('should filter by type', () => {
      const taskLogs = logger.getLogs({ type: 'task' });
      expect(taskLogs).toHaveLength(1);
      expect(taskLogs[0].type).toBe('task');
    });

    it('should filter by level', () => {
      const warningLogs = logger.getLogs({ level: 'warning' });
      expect(warningLogs).toHaveLength(1);
      expect(warningLogs[0].level).toBe('warning');
    });

    it('should filter by search query', () => {
      const searchLogs = logger.getLogs({ searchQuery: 'GitHub' });
      expect(searchLogs).toHaveLength(1);
      expect(searchLogs[0].title).toBe('GitHub Test');
    });

    it('should apply multiple filters', () => {
      const filteredLogs = logger.getLogs({ 
        type: 'task', 
        level: 'info' 
      });
      expect(filteredLogs).toHaveLength(1);
      expect(filteredLogs[0].type).toBe('task');
      expect(filteredLogs[0].level).toBe('info');
    });
  });

  describe('Statistics', () => {
    beforeEach(() => {
      logger.log('connection', 'success', 'Test 1', 'Desc 1');
      logger.log('task', 'info', 'Test 2', 'Desc 2');
      logger.log('github', 'warning', 'Test 3', 'Desc 3');
    });

    it('should return correct statistics', () => {
      const stats = logger.getStatistics();
      
      expect(stats.total).toBe(3);
      expect(stats.byType.connection).toBe(1);
      expect(stats.byType.task).toBe(1);
      expect(stats.byType.github).toBe(1);
      expect(stats.byLevel.success).toBe(1);
      expect(stats.byLevel.info).toBe(1);
      expect(stats.byLevel.warning).toBe(1);
    });
  });

  describe('Export', () => {
    beforeEach(() => {
      logger.log('task', 'success', 'Test Task', 'Test Description');
    });

    it('should export as CSV', () => {
      const csv = logger.exportLogs({ format: 'csv', includeMetadata: false });
      
      expect(csv).toContain('Timestamp,Type,Level,Title,Description');
      expect(csv).toContain('task,success,"Test Task","Test Description"');
    });

    it('should export as JSON', () => {
      const json = logger.exportLogs({ format: 'json', includeMetadata: true });
      const parsed = JSON.parse(json);
      
      expect(Array.isArray(parsed)).toBe(true);
      expect(parsed[0].title).toBe('Test Task');
      expect(parsed[0].type).toBe('task');
    });
  });

  describe('Persistence', () => {
    it('should save logs to localStorage', () => {
      logger.log('task', 'info', 'Test', 'Description');
      
      const stored = localStorage.getItem('workflow_activity_logs');
      expect(stored).toBeTruthy();
      
      const parsed = JSON.parse(stored!);
      expect(parsed).toHaveLength(1);
      expect(parsed[0].title).toBe('Test');
    });

    it('should load logs from localStorage', () => {
      // Manually set localStorage data
      const testLogs = [{
        id: 'test-1',
        timestamp: new Date().toISOString(),
        type: 'task',
        level: 'info',
        title: 'Loaded Task',
        description: 'From storage'
      }];
      
      localStorage.setItem('workflow_activity_logs', JSON.stringify(testLogs));
      
      // Create new logger instance to trigger loading
      (ActivityLogger as any).instance = null;
      const newLogger = ActivityLogger.getInstance();
      
      const logs = newLogger.getLogs();
      expect(logs).toHaveLength(1);
      expect(logs[0].title).toBe('Loaded Task');
    });
  });

  describe('Subscription System', () => {
    it('should notify subscribers when logs are added', () => {
      const mockListener = jest.fn();
      const unsubscribe = logger.subscribe(mockListener);
      
      logger.log('task', 'info', 'New Task', 'Description');
      
      expect(mockListener).toHaveBeenCalledTimes(1);
      expect(mockListener).toHaveBeenCalledWith(expect.arrayContaining([
        expect.objectContaining({ title: 'New Task' })
      ]));
      
      unsubscribe();
    });

    it('should allow unsubscribing', () => {
      const mockListener = jest.fn();
      const unsubscribe = logger.subscribe(mockListener);
      
      unsubscribe();
      logger.log('task', 'info', 'Another Task', 'Description');
      
      expect(mockListener).toHaveBeenCalledTimes(0);
    });
  });

  describe('Clear Logs', () => {
    it('should clear all logs', () => {
      logger.log('task', 'info', 'Task 1', 'Desc 1');
      logger.log('task', 'info', 'Task 2', 'Desc 2');
      
      expect(logger.getLogs()).toHaveLength(2);
      
      logger.clearLogs();
      
      expect(logger.getLogs()).toHaveLength(0);
      expect(localStorage.getItem('workflow_activity_logs')).toBe('[]');
    });
  });
});