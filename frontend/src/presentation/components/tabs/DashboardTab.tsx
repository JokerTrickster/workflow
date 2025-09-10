'use client';

import { useState, useMemo } from 'react';
import { Repository } from '../../../domain/entities/Repository';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select';
import { 
  BarChart3,
  TrendingUp,
  TrendingDown,
  GitBranch,
  GitPullRequest,
  CheckCircle,
  XCircle,
  AlertCircle,
  Clock,
  Users,
  ExternalLink,
  Loader2,
  RefreshCw,
  Calendar,
  Target,
  Activity,
  Download
} from 'lucide-react';
import { useGitHubIssues } from '../../../hooks/useGitHubIssues';
import { useGitHubPullRequests } from '../../../hooks/useGitHubPullRequests';
import { useTasks } from '../../../hooks/useTasks';
import { useGitHubEvents } from '../../../hooks/useGitHubEvents';
import { convertToCSV, downloadCSV, prepareDashboardExportData } from '../../../utils/csvExport';

interface DashboardTabProps {
  repository: Repository;
}

// Task statistics interface
interface TaskStats {
  total: number;
  completed: number;
  inProgress: number;
  pending: number;
  failed: number;
  completionRate: number;
  avgCompletionTime: number; // in hours
  recentActivity: Array<{
    date: string;
    completed: number;
    created: number;
  }>;
}

// Helper to filter data by time range
const filterByTimeRange = (items: any[], timeRange: string): any[] => {
  if (timeRange === 'all') return items;
  
  const now = new Date();
  const days = timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90;
  const cutoff = new Date(now.getTime() - days * 24 * 60 * 60 * 1000);
  
  return items.filter(item => new Date(item.created_at) >= cutoff);
};

// Helper to calculate task statistics from real task data
const calculateTaskStats = (tasks: any[], timeRange: string): TaskStats => {
  const filteredTasks = filterByTimeRange(tasks, timeRange);
  
  const completed = filteredTasks.filter(t => t.status === 'completed').length;
  const inProgress = filteredTasks.filter(t => t.status === 'in_progress').length;
  const pending = filteredTasks.filter(t => t.status === 'pending').length;
  const failed = filteredTasks.filter(t => t.status === 'failed').length;
  const total = filteredTasks.length;
  
  // Calculate completion rate
  const completionRate = total > 0 ? (completed / total) * 100 : 0;
  
  // Calculate average completion time for completed tasks
  const completedTasksWithTimes = filteredTasks.filter(t => 
    t.status === 'completed' && t.completed_at && t.started_at
  );
  
  const avgCompletionTime = completedTasksWithTimes.length > 0 
    ? completedTasksWithTimes.reduce((acc, task) => {
        const start = new Date(task.started_at).getTime();
        const end = new Date(task.completed_at).getTime();
        return acc + (end - start) / (1000 * 60 * 60); // hours
      }, 0) / completedTasksWithTimes.length
    : 0;
  
  // Generate recent activity (last 7 days)
  const recentActivity = Array.from({ length: 7 }, (_, i) => {
    const date = new Date(Date.now() - i * 24 * 60 * 60 * 1000);
    const dateStr = date.toISOString().split('T')[0];
    
    const dayTasks = filteredTasks.filter(task => {
      const taskDate = new Date(task.created_at).toISOString().split('T')[0];
      return taskDate === dateStr;
    });
    
    const dayCompleted = filteredTasks.filter(task => {
      if (!task.completed_at) return false;
      const completedDate = new Date(task.completed_at).toISOString().split('T')[0];
      return completedDate === dateStr;
    });
    
    return {
      date: dateStr,
      completed: dayCompleted.length,
      created: dayTasks.length
    };
  }).reverse();
  
  return {
    total,
    completed,
    inProgress,
    pending,
    failed,
    completionRate,
    avgCompletionTime,
    recentActivity
  };
};

// Progress bar component
const ProgressBar = ({ value, max, className = '' }: { value: number; max: number; className?: string }) => {
  const percentage = Math.round((value / max) * 100);
  return (
    <div className={`w-full bg-gray-200 rounded-full h-2 ${className}`}>
      <div 
        className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
        style={{ width: `${percentage}%` }}
      />
    </div>
  );
};

// Metric card component
const MetricCard = ({ 
  title, 
  value, 
  subtitle, 
  icon: Icon, 
  trend, 
  color = 'text-gray-600',
  loading = false 
}: {
  title: string;
  value: string | number;
  subtitle: string;
  icon: React.ElementType;
  trend?: 'up' | 'down' | 'neutral';
  color?: string;
  loading?: boolean;
}) => (
  <Card>
    <CardHeader className="pb-2">
      <CardTitle className="text-sm font-medium flex items-center justify-between">
        <span>{title}</span>
        <Icon className={`h-4 w-4 ${color}`} />
      </CardTitle>
    </CardHeader>
    <CardContent>
      {loading ? (
        <div className="flex items-center space-x-2">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span className="text-sm text-muted-foreground">Loading...</span>
        </div>
      ) : (
        <>
          <div className="text-2xl font-bold">{value}</div>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <span>{subtitle}</span>
            {trend && (
              <span className={`flex items-center ${
                trend === 'up' ? 'text-green-600' : 
                trend === 'down' ? 'text-red-600' : 
                'text-gray-600'
              }`}>
                {trend === 'up' ? <TrendingUp className="h-3 w-3 ml-1" /> : 
                 trend === 'down' ? <TrendingDown className="h-3 w-3 ml-1" /> : null}
              </span>
            )}
          </div>
        </>
      )}
    </CardContent>
  </Card>
);

export function DashboardTab({ repository }: DashboardTabProps) {
  const [timeRange, setTimeRange] = useState<string>('30d');
  
  // Get GitHub data
  const { 
    data: issuesData, 
    isLoading: issuesLoading, 
    error: issuesError,
    refetch: refetchIssues 
  } = useGitHubIssues({
    repoId: repository.full_name,
    params: { state: 'all', per_page: 100 }
  });

  const { 
    data: prsData, 
    isLoading: prsLoading, 
    error: prsError,
    refetch: refetchPRs 
  } = useGitHubPullRequests({
    repoId: repository.full_name,
    params: { state: 'all', per_page: 100 }
  });

  // Get local task data
  const {
    data: tasksData,
    isLoading: tasksLoading,
    error: tasksError,
    refetch: refetchTasks
  } = useTasks({
    repositoryId: repository.id,
    enabled: true
  });

  // Get GitHub events for activity timeline
  const {
    data: eventsData,
    isLoading: eventsLoading,
    error: eventsError,
    refetch: refetchEvents
  } = useGitHubEvents({
    repoId: repository.full_name,
    enabled: true
  });

  // Calculate task statistics from real data
  const taskStats = useMemo(() => {
    if (!tasksData) {
      return {
        total: 0,
        completed: 0,
        inProgress: 0,
        pending: 0,
        failed: 0,
        completionRate: 0,
        avgCompletionTime: 0,
        recentActivity: []
      };
    }
    return calculateTaskStats(tasksData, timeRange);
  }, [tasksData, timeRange]);

  // Calculate GitHub statistics with time range filtering
  const githubStats = useMemo(() => {
    if (!issuesData?.issues || !prsData?.pullRequests) {
      return null;
    }

    const allIssues = issuesData.issues;
    const allPRs = prsData.pullRequests;
    
    // Filter by time range for statistics
    const filteredIssues = filterByTimeRange(allIssues, timeRange);
    const filteredPRs = filterByTimeRange(allPRs, timeRange);

    const openIssues = filteredIssues.filter(issue => issue.state === 'open').length;
    const closedIssues = filteredIssues.filter(issue => issue.state === 'closed').length;
    const openPRs = filteredPRs.filter(pr => pr.state === 'open').length;
    const closedPRs = filteredPRs.filter(pr => pr.state === 'closed').length;
    const mergedPRs = filteredPRs.filter(pr => pr.merged).length;

    // Calculate recent activity based on current time range
    const days = timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : timeRange === '90d' ? 90 : 30;
    const cutoffDate = new Date(Date.now() - days * 24 * 60 * 60 * 1000);
    const recentIssues = allIssues.filter(issue => 
      new Date(issue.created_at) >= cutoffDate
    ).length;
    const recentPRs = allPRs.filter(pr => 
      new Date(pr.created_at) >= cutoffDate
    ).length;

    // Calculate trends (compare to previous period)
    const prevCutoff = new Date(cutoffDate.getTime() - days * 24 * 60 * 60 * 1000);
    const prevIssues = allIssues.filter(issue => {
      const date = new Date(issue.created_at);
      return date >= prevCutoff && date < cutoffDate;
    }).length;
    const prevPRs = allPRs.filter(pr => {
      const date = new Date(pr.created_at);
      return date >= prevCutoff && date < cutoffDate;
    }).length;

    return {
      issues: {
        open: openIssues,
        closed: closedIssues,
        total: filteredIssues.length,
        allTimeTotal: allIssues.length,
        recent: recentIssues,
        trend: recentIssues > prevIssues ? 'up' : recentIssues < prevIssues ? 'down' : 'neutral'
      },
      pullRequests: {
        open: openPRs,
        closed: closedPRs,
        merged: mergedPRs,
        total: filteredPRs.length,
        allTimeTotal: allPRs.length,
        recent: recentPRs,
        mergeRate: filteredPRs.length > 0 ? Math.round((mergedPRs / filteredPRs.length) * 100) : 0,
        trend: recentPRs > prevPRs ? 'up' : recentPRs < prevPRs ? 'down' : 'neutral'
      }
    };
  }, [issuesData, prsData, timeRange]);

  const handleRefresh = () => {
    refetchIssues();
    refetchPRs();
    refetchTasks();
    refetchEvents();
  };

  const handleExport = () => {
    try {
      const exportData = prepareDashboardExportData(
        repository,
        githubStats,
        taskStats,
        timeRange
      );
      
      const csvContent = convertToCSV(exportData);
      const filename = `dashboard-${repository.name}-${timeRange}-${new Date().toISOString().split('T')[0]}.csv`;
      
      downloadCSV(csvContent, filename);
    } catch (error) {
      console.error('Failed to export dashboard data:', error);
      // Could add toast notification here
    }
  };

  const hasError = issuesError || prsError || tasksError || eventsError;
  const isLoading = issuesLoading || prsLoading || tasksLoading || eventsLoading;

  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 overflow-auto">
        <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <BarChart3 className="h-5 w-5" />
          Dashboard
        </h2>
        <div className="flex items-center gap-2">
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="7d">Last 7 days</SelectItem>
              <SelectItem value="30d">Last 30 days</SelectItem>
              <SelectItem value="90d">Last 90 days</SelectItem>
              <SelectItem value="all">All time</SelectItem>
            </SelectContent>
          </Select>
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleRefresh}
            disabled={isLoading}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleExport}
            disabled={isLoading || hasError}
          >
            <Download className="h-4 w-4 mr-2" />
            Export CSV
          </Button>
        </div>
      </div>

      {/* Error State */}
      {hasError && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="flex items-center justify-between p-4">
            <div className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-500" />
              <div>
                <p className="font-medium text-red-900">Failed to load GitHub data</p>
                <p className="text-sm text-red-700">
                  {issuesError?.message || prsError?.message || 'Unknown error occurred'}
                </p>
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={handleRefresh}>
              Try again
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Repository Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>Repository Overview</span>
            <Button variant="ghost" size="sm" asChild>
              <a href={repository.html_url} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="h-4 w-4 mr-2" />
                View on GitHub
              </a>
            </Button>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-yellow-600">{repository.stargazers_count}</div>
              <div className="text-xs text-muted-foreground">Stars</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">{repository.forks_count}</div>
              <div className="text-xs text-muted-foreground">Forks</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {repository.language || 'N/A'}
              </div>
              <div className="text-xs text-muted-foreground">Language</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">
                {new Date(repository.updated_at).toLocaleDateString()}
              </div>
              <div className="text-xs text-muted-foreground">Last Updated</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* GitHub Issues */}
        <MetricCard
          title="Open Issues"
          value={isLoading ? '-' : githubStats?.issues.open || 0}
          subtitle={`${githubStats?.issues.allTimeTotal || 0} total`}
          icon={AlertCircle}
          trend={githubStats?.issues.trend || 'neutral'}
          color="text-red-500"
          loading={isLoading}
        />

        {/* GitHub Pull Requests */}
        <MetricCard
          title="Open PRs"
          value={isLoading ? '-' : githubStats?.pullRequests.open || 0}
          subtitle={`${githubStats?.pullRequests.mergeRate || 0}% merge rate`}
          icon={GitPullRequest}
          trend={githubStats?.pullRequests.trend || 'neutral'}
          color="text-purple-500"
          loading={isLoading}
        />

        {/* Local Tasks */}
        <MetricCard
          title="Task Progress"
          value={`${taskStats.completed}/${taskStats.total}`}
          subtitle={`${Math.round(taskStats.completionRate)}% completed`}
          icon={Target}
          trend={taskStats.completionRate > 50 ? 'up' : taskStats.completionRate > 25 ? 'neutral' : 'down'}
          color="text-green-500"
          loading={tasksLoading}
        />

        {/* Average Time */}
        <MetricCard
          title="Avg Completion"
          value={taskStats.avgCompletionTime > 0 ? `${taskStats.avgCompletionTime.toFixed(1)}h` : '-'}
          subtitle="per task"
          icon={Clock}
          trend={taskStats.avgCompletionTime < 8 ? 'up' : taskStats.avgCompletionTime < 24 ? 'neutral' : 'down'}
          color="text-blue-500"
          loading={tasksLoading}
        />
      </div>

      {/* Detailed Statistics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* GitHub Activity */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <GitBranch className="h-5 w-5 text-purple-500" />
              GitHub Activity
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : githubStats ? (
              <>
                {/* Issues Section */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Issues ({timeRange})</span>
                    <Badge variant="outline">{githubStats.issues.total} in period</Badge>
                  </div>
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full bg-red-500"></div>
                        Open
                      </span>
                      <span>{githubStats.issues.open}</span>
                    </div>
                    <ProgressBar 
                      value={githubStats.issues.open} 
                      max={githubStats.issues.total} 
                      className="bg-red-200"
                    />
                  </div>
                  <div className="space-y-2 mt-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full bg-green-500"></div>
                        Closed
                      </span>
                      <span>{githubStats.issues.closed}</span>
                    </div>
                    <ProgressBar 
                      value={githubStats.issues.closed} 
                      max={githubStats.issues.total}
                      className="bg-green-200"
                    />
                  </div>
                </div>

                {/* Pull Requests Section */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Pull Requests ({timeRange})</span>
                    <Badge variant="outline">{githubStats.pullRequests.total} in period</Badge>
                  </div>
                  <div className="grid grid-cols-3 gap-2 text-xs">
                    <div className="text-center p-2 bg-yellow-50 rounded">
                      <div className="font-semibold text-yellow-600">{githubStats.pullRequests.open}</div>
                      <div className="text-yellow-800">Open</div>
                    </div>
                    <div className="text-center p-2 bg-green-50 rounded">
                      <div className="font-semibold text-green-600">{githubStats.pullRequests.merged}</div>
                      <div className="text-green-800">Merged</div>
                    </div>
                    <div className="text-center p-2 bg-gray-50 rounded">
                      <div className="font-semibold">{githubStats.pullRequests.closed}</div>
                      <div className="text-gray-600">Closed</div>
                    </div>
                  </div>
                </div>

                {/* Recent Activity */}
                <div>
                  <span className="text-sm font-medium">Recent Activity ({timeRange})</span>
                  <div className="mt-2 grid grid-cols-2 gap-4 text-sm">
                    <div className="flex items-center gap-2">
                      <AlertCircle className="h-4 w-4 text-red-500" />
                      <span>{githubStats.issues.recent} new issues</span>
                      {githubStats.issues.trend === 'up' && <TrendingUp className="h-3 w-3 text-green-500" />}
                      {githubStats.issues.trend === 'down' && <TrendingDown className="h-3 w-3 text-red-500" />}
                    </div>
                    <div className="flex items-center gap-2">
                      <GitPullRequest className="h-4 w-4 text-purple-500" />
                      <span>{githubStats.pullRequests.recent} new PRs</span>
                      {githubStats.pullRequests.trend === 'up' && <TrendingUp className="h-3 w-3 text-green-500" />}
                      {githubStats.pullRequests.trend === 'down' && <TrendingDown className="h-3 w-3 text-red-500" />}
                    </div>
                  </div>
                </div>
              </>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <GitBranch className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>No GitHub data available</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Task Progress */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircle className="h-5 w-5 text-green-500" />
              Task Progress
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Task Status Breakdown */}
            <div className="space-y-3">
              <div className="flex items-center justify-between text-sm">
                <span className="flex items-center gap-2">
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  Completed
                </span>
                <span className="font-medium">{taskStats.completed}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-blue-500" />
                  In Progress
                </span>
                <span className="font-medium">{taskStats.inProgress}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="flex items-center gap-2">
                  <AlertCircle className="h-4 w-4 text-yellow-500" />
                  Pending
                </span>
                <span className="font-medium">{taskStats.pending}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="flex items-center gap-2">
                  <XCircle className="h-4 w-4 text-red-500" />
                  Failed
                </span>
                <span className="font-medium">{taskStats.failed}</span>
              </div>
            </div>

            {/* Progress Visualization */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium">Overall Progress</span>
                <span className="text-sm text-muted-foreground">
                  {Math.round(taskStats.completionRate)}%
                </span>
              </div>
              <ProgressBar 
                value={taskStats.completed} 
                max={taskStats.total}
                className="bg-gray-200"
              />
            </div>

            {/* Recent Activity Chart */}
            <div>
              <span className="text-sm font-medium">Recent Activity</span>
              <div className="mt-2 grid grid-cols-7 gap-1">
                {taskStats.recentActivity.map((day, index) => (
                  <div key={index} className="text-center">
                    <div className="text-xs text-muted-foreground mb-1">
                      {new Date(day.date).getDate()}
                    </div>
                    <div className="space-y-1">
                      <div 
                        className="bg-green-200 rounded-sm" 
                        style={{ height: `${Math.max(2, day.completed * 4)}px` }}
                        title={`${day.completed} completed`}
                      />
                      <div 
                        className="bg-blue-200 rounded-sm" 
                        style={{ height: `${Math.max(2, day.created * 4)}px` }}
                        title={`${day.created} created`}
                      />
                    </div>
                  </div>
                ))}
              </div>
              <div className="flex items-center justify-center gap-4 mt-2 text-xs text-muted-foreground">
                <div className="flex items-center gap-1">
                  <div className="w-3 h-2 bg-green-200 rounded-sm"></div>
                  <span>Completed</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="w-3 h-2 bg-blue-200 rounded-sm"></div>
                  <span>Created</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Activity Timeline */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5 text-blue-500" />
            Recent Activity Timeline
          </CardTitle>
        </CardHeader>
        <CardContent>
          {eventsLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : eventsData && eventsData.length > 0 ? (
            <div className="space-y-4 max-h-80 overflow-y-auto">
              {eventsData.slice(0, 10).map((event, index) => {
                const getEventInfo = (event: any) => {
                  switch (event.type) {
                    case 'PushEvent':
                      return {
                        icon: <GitBranch className="h-4 w-4 text-green-500" />,
                        action: 'pushed',
                        details: `${event.payload.commits?.length || 0} commit(s) to ${event.payload.ref?.replace('refs/heads/', '') || 'unknown branch'}`,
                        color: 'text-green-600'
                      };
                    case 'PullRequestEvent':
                      return {
                        icon: <GitPullRequest className="h-4 w-4 text-purple-500" />,
                        action: event.payload.action === 'opened' ? 'opened PR' : 
                               event.payload.action === 'closed' ? 'closed PR' : 
                               event.payload.action === 'merged' ? 'merged PR' : 'updated PR',
                        details: event.payload.pull_request?.title || 'Pull request',
                        color: 'text-purple-600'
                      };
                    case 'IssuesEvent':
                      return {
                        icon: <AlertCircle className="h-4 w-4 text-red-500" />,
                        action: event.payload.action === 'opened' ? 'opened issue' : 
                               event.payload.action === 'closed' ? 'closed issue' : 'updated issue',
                        details: event.payload.issue?.title || 'Issue',
                        color: 'text-red-600'
                      };
                    case 'CreateEvent':
                      return {
                        icon: <GitBranch className="h-4 w-4 text-blue-500" />,
                        action: `created ${event.payload.ref_type}`,
                        details: event.payload.ref || event.payload.ref_type,
                        color: 'text-blue-600'
                      };
                    case 'DeleteEvent':
                      return {
                        icon: <XCircle className="h-4 w-4 text-gray-500" />,
                        action: `deleted ${event.payload.ref_type}`,
                        details: event.payload.ref || event.payload.ref_type,
                        color: 'text-gray-600'
                      };
                    default:
                      return {
                        icon: <Activity className="h-4 w-4 text-gray-500" />,
                        action: event.type.replace('Event', '').toLowerCase(),
                        details: '',
                        color: 'text-gray-600'
                      };
                  }
                };

                const eventInfo = getEventInfo(event);
                const timeAgo = new Date(event.created_at).toLocaleString();

                return (
                  <div key={event.id} className="flex items-start space-x-3 pb-3 border-b border-gray-100 last:border-b-0">
                    <div className="flex-shrink-0 mt-1">
                      {eventInfo.icon}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 text-sm">
                        <span className="font-medium">{event.actor.login}</span>
                        <span className={eventInfo.color}>{eventInfo.action}</span>
                      </div>
                      {eventInfo.details && (
                        <div className="text-sm text-muted-foreground mt-1 truncate">
                          {eventInfo.details}
                        </div>
                      )}
                      <div className="text-xs text-muted-foreground mt-1">
                        {timeAgo}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : eventsError ? (
            <div className="text-center py-8 text-muted-foreground">
              <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>Failed to load activity timeline</p>
              <p className="text-xs">{eventsError?.message}</p>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>No recent activity</p>
            </div>
          )}
        </CardContent>
      </Card>
        </div>
      </div>
    </div>
  );
}