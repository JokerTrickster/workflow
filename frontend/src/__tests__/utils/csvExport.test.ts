import { convertToCSV, prepareDashboardExportData } from '../../utils/csvExport';

describe('csvExport', () => {
  const mockRepository = {
    id: 1,
    name: 'test-repo',
    full_name: 'user/test-repo',
    html_url: 'https://github.com/user/test-repo',
    language: 'TypeScript',
    stargazers_count: 100,
    forks_count: 20,
    updated_at: '2024-01-01T00:00:00Z',
  };

  const mockGithubStats = {
    issues: {
      open: 5,
      closed: 10,
      total: 15,
      recent: 2,
    },
    pullRequests: {
      open: 3,
      closed: 7,
      merged: 6,
      total: 10,
      recent: 1,
      mergeRate: 60,
    },
  };

  const mockTaskStats = {
    total: 20,
    completed: 15,
    inProgress: 3,
    pending: 2,
    failed: 0,
    completionRate: 75,
    avgCompletionTime: 4.5,
  };

  describe('prepareDashboardExportData', () => {
    it('should prepare export data correctly', () => {
      const result = prepareDashboardExportData(
        mockRepository,
        mockGithubStats,
        mockTaskStats,
        '30d'
      );

      expect(result.repository.name).toBe('test-repo');
      expect(result.repository.fullName).toBe('user/test-repo');
      expect(result.github.issues.open).toBe(5);
      expect(result.github.pullRequests.mergeRate).toBe(60);
      expect(result.tasks.completionRate).toBe(75);
      expect(result.timeRange).toBe('30d');
      expect(result.exportedAt).toBeDefined();
    });

    it('should handle missing github stats', () => {
      const result = prepareDashboardExportData(
        mockRepository,
        null,
        mockTaskStats,
        '7d'
      );

      expect(result.github.issues.open).toBe(0);
      expect(result.github.pullRequests.mergeRate).toBe(0);
    });
  });

  describe('convertToCSV', () => {
    it('should convert data to CSV format', () => {
      const exportData = prepareDashboardExportData(
        mockRepository,
        mockGithubStats,
        mockTaskStats,
        '30d'
      );

      const csv = convertToCSV(exportData);

      expect(csv).toContain('Metric,Category,Value,Details');
      expect(csv).toContain('Repository Name,Repository,test-repo,');
      expect(csv).toContain('Open Issues,GitHub Issues,5,30d period');
      expect(csv).toContain('Merge Rate,GitHub PRs,60%,30d period');
      expect(csv).toContain('Completion Rate,Local Tasks,75.0%,30d period');
    });

    it('should escape quotes in CSV data', () => {
      const testData = {
        ...prepareDashboardExportData(mockRepository, mockGithubStats, mockTaskStats, '30d'),
        repository: {
          ...prepareDashboardExportData(mockRepository, mockGithubStats, mockTaskStats, '30d').repository,
          name: 'test "quoted" repo',
        },
      };

      const csv = convertToCSV(testData);
      
      expect(csv).toContain('"test ""quoted"" repo"');
    });
  });
});