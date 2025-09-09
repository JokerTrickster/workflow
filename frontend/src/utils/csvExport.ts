/**
 * Utility functions for exporting dashboard data to CSV format
 */

export interface DashboardExportData {
  repository: {
    name: string;
    fullName: string;
    url: string;
    language: string;
    stars: number;
    forks: number;
    lastUpdated: string;
  };
  github: {
    issues: {
      open: number;
      closed: number;
      total: number;
      recent: number;
    };
    pullRequests: {
      open: number;
      closed: number;
      merged: number;
      total: number;
      recent: number;
      mergeRate: number;
    };
  };
  tasks: {
    total: number;
    completed: number;
    inProgress: number;
    pending: number;
    failed: number;
    completionRate: number;
    avgCompletionTime: number;
  };
  timeRange: string;
  exportedAt: string;
}

/**
 * Convert data to CSV format
 */
export function convertToCSV(data: DashboardExportData): string {
  const rows = [
    // Header
    ['Metric', 'Category', 'Value', 'Details'],
    
    // Repository Info
    ['Repository Name', 'Repository', data.repository.name, ''],
    ['Full Name', 'Repository', data.repository.fullName, ''],
    ['Language', 'Repository', data.repository.language, ''],
    ['Stars', 'Repository', data.repository.stars.toString(), ''],
    ['Forks', 'Repository', data.repository.forks.toString(), ''],
    ['Last Updated', 'Repository', data.repository.lastUpdated, ''],
    
    // GitHub Issues
    ['Open Issues', 'GitHub Issues', data.github.issues.open.toString(), `${data.timeRange} period`],
    ['Closed Issues', 'GitHub Issues', data.github.issues.closed.toString(), `${data.timeRange} period`],
    ['Total Issues', 'GitHub Issues', data.github.issues.total.toString(), `${data.timeRange} period`],
    ['Recent Issues', 'GitHub Issues', data.github.issues.recent.toString(), `${data.timeRange} period`],
    
    // GitHub Pull Requests
    ['Open PRs', 'GitHub PRs', data.github.pullRequests.open.toString(), `${data.timeRange} period`],
    ['Closed PRs', 'GitHub PRs', data.github.pullRequests.closed.toString(), `${data.timeRange} period`],
    ['Merged PRs', 'GitHub PRs', data.github.pullRequests.merged.toString(), `${data.timeRange} period`],
    ['Total PRs', 'GitHub PRs', data.github.pullRequests.total.toString(), `${data.timeRange} period`],
    ['Recent PRs', 'GitHub PRs', data.github.pullRequests.recent.toString(), `${data.timeRange} period`],
    ['Merge Rate', 'GitHub PRs', `${data.github.pullRequests.mergeRate}%`, `${data.timeRange} period`],
    
    // Local Tasks
    ['Total Tasks', 'Local Tasks', data.tasks.total.toString(), `${data.timeRange} period`],
    ['Completed Tasks', 'Local Tasks', data.tasks.completed.toString(), `${data.timeRange} period`],
    ['In Progress Tasks', 'Local Tasks', data.tasks.inProgress.toString(), `${data.timeRange} period`],
    ['Pending Tasks', 'Local Tasks', data.tasks.pending.toString(), `${data.timeRange} period`],
    ['Failed Tasks', 'Local Tasks', data.tasks.failed.toString(), `${data.timeRange} period`],
    ['Completion Rate', 'Local Tasks', `${data.tasks.completionRate.toFixed(1)}%`, `${data.timeRange} period`],
    ['Avg Completion Time', 'Local Tasks', `${data.tasks.avgCompletionTime.toFixed(1)}h`, `${data.timeRange} period`],
    
    // Metadata
    ['Time Range', 'Metadata', data.timeRange, ''],
    ['Exported At', 'Metadata', data.exportedAt, ''],
  ];

  return rows.map(row => 
    row.map(cell => `"${cell.replace(/"/g, '""')}"`).join(',')
  ).join('\n');
}

/**
 * Download CSV file
 */
export function downloadCSV(csvContent: string, filename: string): void {
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  
  if (link.download !== undefined) {
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }
}

/**
 * Generate dashboard export data
 */
export function prepareDashboardExportData(
  repository: any,
  githubStats: any,
  taskStats: any,
  timeRange: string
): DashboardExportData {
  return {
    repository: {
      name: repository.name,
      fullName: repository.full_name,
      url: repository.html_url,
      language: repository.language || 'N/A',
      stars: repository.stargazers_count,
      forks: repository.forks_count,
      lastUpdated: new Date(repository.updated_at).toLocaleDateString(),
    },
    github: {
      issues: {
        open: githubStats?.issues.open || 0,
        closed: githubStats?.issues.closed || 0,
        total: githubStats?.issues.total || 0,
        recent: githubStats?.issues.recent || 0,
      },
      pullRequests: {
        open: githubStats?.pullRequests.open || 0,
        closed: githubStats?.pullRequests.closed || 0,
        merged: githubStats?.pullRequests.merged || 0,
        total: githubStats?.pullRequests.total || 0,
        recent: githubStats?.pullRequests.recent || 0,
        mergeRate: githubStats?.pullRequests.mergeRate || 0,
      },
    },
    tasks: {
      total: taskStats.total,
      completed: taskStats.completed,
      inProgress: taskStats.inProgress,
      pending: taskStats.pending,
      failed: taskStats.failed,
      completionRate: taskStats.completionRate,
      avgCompletionTime: taskStats.avgCompletionTime,
    },
    timeRange,
    exportedAt: new Date().toISOString(),
  };
}