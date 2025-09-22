import React, { useEffect, useCallback, useState } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { useTaskHistory } from '../../hooks/useTaskHistory';
import { TaskHistoryItem } from './TaskHistoryItem';
import { Pagination } from './Pagination';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import {
  History,
  Search,
  RefreshCw,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
  Filter
} from 'lucide-react';

interface TaskHistoryProps {
  repository: Repository;
}

/**
 * Main TaskHistory component that displays paginated workflow history for a repository
 */
export function TaskHistory({ repository }: TaskHistoryProps) {
  const {
    tasks,
    pagination,
    isLoading,
    error,
    loadTaskHistory,
    loadNextPage,
    loadPrevPage,
    loadPage,
    refresh,
    clearError,
    hasNextPage,
    hasPrevPage,
    currentPage,
    totalPages,
  } = useTaskHistory();

  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'>('all');
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Filter tasks based on search and status
  const filteredTasks = tasks.filter(task => {
    const matchesSearch = !searchQuery ||
      task.tasks.toLowerCase().includes(searchQuery.toLowerCase()) ||
      task.request_id.toLowerCase().includes(searchQuery.toLowerCase());

    const matchesStatus = statusFilter === 'all' || task.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  // Initial load when repository changes
  useEffect(() => {
    if (repository.name) {
      loadTaskHistory(repository.name, { page: 1, limit: 20 });
    }
  }, [repository.name, loadTaskHistory]);

  // Auto-refresh polling for active tasks
  useEffect(() => {
    if (!autoRefresh) return;

    const hasActiveTasks = tasks.some(task =>
      task.status === 'pending' || task.status === 'processing'
    );

    if (!hasActiveTasks) return;

    const intervalId = setInterval(() => {
      refresh();
    }, 5000); // Poll every 5 seconds

    return () => clearInterval(intervalId);
  }, [autoRefresh, tasks, refresh]);

  const handleRefresh = useCallback(() => {
    refresh();
  }, [refresh]);

  const handlePageChange = useCallback((page: number) => {
    loadPage(page);
  }, [loadPage]);

  const getStatusStats = () => {
    const stats = {
      total: tasks.length,
      completed: tasks.filter(t => t.status === 'completed').length,
      failed: tasks.filter(t => t.status === 'failed').length,
      processing: tasks.filter(t => t.status === 'processing').length,
      pending: tasks.filter(t => t.status === 'pending').length,
    };
    return stats;
  };

  const stats = getStatusStats();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <History className="h-5 w-5" />
            Task History
          </h2>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>{stats.total} total</span>
            {stats.processing > 0 && (
              <>
                <span>•</span>
                <Badge variant="default" className="bg-blue-100 text-blue-800">
                  {stats.processing} running
                </Badge>
              </>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={autoRefresh ? 'bg-green-50 text-green-700' : ''}
          >
            <Clock className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-pulse' : ''}`} />
            Auto-refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isLoading}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <div className="p-2 bg-gray-100 rounded-lg">
                <History className="h-4 w-4 text-gray-600" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Total</p>
                <p className="text-lg font-semibold">{stats.total}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <div className="p-2 bg-green-100 rounded-lg">
                <CheckCircle className="h-4 w-4 text-green-600" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Completed</p>
                <p className="text-lg font-semibold">{stats.completed}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Loader2 className="h-4 w-4 text-blue-600" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Processing</p>
                <p className="text-lg font-semibold">{stats.processing}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <div className="p-2 bg-red-100 rounded-lg">
                <XCircle className="h-4 w-4 text-red-600" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Failed</p>
                <p className="text-lg font-semibold">{stats.failed}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <div className="p-2 bg-yellow-100 rounded-lg">
                <Clock className="h-4 w-4 text-yellow-600" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Pending</p>
                <p className="text-lg font-semibold">{stats.pending}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <Filter className="h-4 w-4" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search tasks or request ID..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Status Filter */}
            <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as any)}>
              <SelectTrigger>
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="processing">Processing</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
                <SelectItem value="failed">Failed</SelectItem>
                <SelectItem value="cancelled">Cancelled</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Error State */}
      {error && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="p-4">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-red-500 mt-0.5" />
              <div className="flex-1">
                <h3 className="font-medium text-red-800">Failed to load task history</h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
                <div className="flex gap-2 mt-3">
                  <Button variant="outline" size="sm" onClick={clearError}>
                    Dismiss
                  </Button>
                  <Button variant="outline" size="sm" onClick={handleRefresh}>
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Retry
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Loading State */}
      {isLoading && tasks.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin mb-4" />
          <h3 className="text-lg font-semibold mb-2">Loading Task History</h3>
          <p className="text-sm text-muted-foreground text-center max-w-md">
            Fetching workflow history for {repository.name}...
          </p>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && filteredTasks.length === 0 && tasks.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12">
          <History className="h-16 w-16 mb-4 opacity-50" />
          <h3 className="text-lg font-semibold mb-2">No Task History</h3>
          <p className="text-sm text-muted-foreground text-center max-w-md">
            No tasks have been executed for this repository yet.
            Task history will appear here after Claude tasks are run.
          </p>
        </div>
      )}

      {/* Filtered Empty State */}
      {!isLoading && !error && filteredTasks.length === 0 && tasks.length > 0 && (
        <div className="flex flex-col items-center justify-center py-12">
          <Search className="h-16 w-16 mb-4 opacity-50" />
          <h3 className="text-lg font-semibold mb-2">No Matching Tasks</h3>
          <p className="text-sm text-muted-foreground text-center max-w-md">
            No tasks match your current filters. Try adjusting your search criteria.
          </p>
          <Button
            variant="outline"
            className="mt-4"
            onClick={() => {
              setSearchQuery('');
              setStatusFilter('all');
            }}
          >
            Clear Filters
          </Button>
        </div>
      )}

      {/* Task List */}
      {!isLoading && filteredTasks.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>
              Showing {filteredTasks.length} of {tasks.length} tasks
              {pagination && ` (Page ${currentPage} of ${totalPages})`}
            </span>
          </div>

          <div className="space-y-3">
            {filteredTasks.map((task) => (
              <TaskHistoryItem key={task.request_id} task={task} />
            ))}
          </div>

          {/* Pagination */}
          {pagination && totalPages > 1 && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              hasNextPage={hasNextPage}
              hasPrevPage={hasPrevPage}
              isLoading={isLoading}
              onPageChange={handlePageChange}
              onNextPage={loadNextPage}
              onPrevPage={loadPrevPage}
            />
          )}
        </div>
      )}
    </div>
  );
}