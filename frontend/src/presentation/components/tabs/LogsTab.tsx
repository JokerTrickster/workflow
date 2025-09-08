'use client';

import { useState, useMemo, useEffect } from 'react';
import { Repository } from '../../../domain/entities/Repository';
import { ActivityLogger } from '../../../services/ActivityLogger';
import { ActivityLog, ActivityType, ActivityLevel } from '../../../types/activity';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { Input } from '../../../components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select';
import { Separator } from '../../../components/ui/separator';
import { 
  Activity, 
  Search, 
  Filter, 
  Download, 
  Calendar,
  GitBranch,
  GitPullRequest,
  CheckCircle,
  XCircle,
  AlertCircle,
  Clock,
  Users,
  ExternalLink,
  Trash2,
  BarChart3
} from 'lucide-react';

interface LogsTabProps {
  repository: Repository;
}


const formatRelativeTime = (timestamp: string): string => {
  const now = new Date();
  const time = new Date(timestamp);
  const diffInMinutes = Math.floor((now.getTime() - time.getTime()) / (1000 * 60));
  
  if (diffInMinutes < 1) return 'Just now';
  if (diffInMinutes < 60) return `${diffInMinutes} min ago`;
  
  const diffInHours = Math.floor(diffInMinutes / 60);
  if (diffInHours < 24) return `${diffInHours} hour${diffInHours > 1 ? 's' : ''} ago`;
  
  const diffInDays = Math.floor(diffInHours / 24);
  return `${diffInDays} day${diffInDays > 1 ? 's' : ''} ago`;
};

const getActivityIcon = (type: ActivityType, level: ActivityLevel) => {
  const baseClasses = "h-4 w-4";
  
  switch (type) {
    case 'connection':
      return level === 'success' 
        ? <CheckCircle className={`${baseClasses} text-green-500`} />
        : <XCircle className={`${baseClasses} text-red-500`} />;
    case 'task':
      return level === 'success' 
        ? <CheckCircle className={`${baseClasses} text-green-500`} />
        : <Clock className={`${baseClasses} text-blue-500`} />;
    case 'github':
      return level === 'warning'
        ? <AlertCircle className={`${baseClasses} text-yellow-500`} />
        : <GitBranch className={`${baseClasses} text-purple-500`} />;
    case 'navigation':
      return <Users className={`${baseClasses} text-gray-500`} />;
    default:
      return <Activity className={`${baseClasses} text-gray-500`} />;
  }
};

const getActivityBadgeColor = (type: ActivityType): string => {
  switch (type) {
    case 'connection': return 'bg-green-100 text-green-800';
    case 'task': return 'bg-blue-100 text-blue-800';
    case 'github': return 'bg-purple-100 text-purple-800';
    case 'navigation': return 'bg-gray-100 text-gray-800';
    default: return 'bg-gray-100 text-gray-800';
  }
};

export function LogsTab({ repository }: LogsTabProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState<ActivityType | 'all'>('all');
  const [selectedDateRange, setSelectedDateRange] = useState<string>('24h');
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [stats, setStats] = useState<{
    total: number;
    recentActivity: number;
  } | null>(null);
  
  const activityLogger = ActivityLogger.getInstance();

  // Subscribe to activity logger updates
  useEffect(() => {
    const unsubscribe = activityLogger.subscribe((updatedLogs) => {
      setLogs(updatedLogs);
    });

    // Initial load
    setLogs(activityLogger.getLogs());
    setStats(activityLogger.getStatistics());

    // Log workspace access
    activityLogger.logWorkspaceAccess(repository.name, 'opened');

    return () => {
      unsubscribe();
    };
  }, [repository.id, repository.name, activityLogger]);
  
  // Filter logs based on search and filters
  const filteredLogs = useMemo(() => {
    let filtered = logs;
    
    // Filter by repository if needed (for multi-repo support)
    filtered = filtered.filter(log => 
      !log.metadata?.repositoryId || 
      log.metadata.repositoryId === repository.id ||
      log.metadata?.repositoryName === repository.name
    );
    
    // Filter by type
    if (selectedType !== 'all') {
      filtered = filtered.filter(log => log.type === selectedType);
    }
    
    // Filter by date range
    const now = new Date();
    let cutoffTime: Date;
    
    switch (selectedDateRange) {
      case '1h':
        cutoffTime = new Date(now.getTime() - 60 * 60 * 1000);
        break;
      case '24h':
        cutoffTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        break;
      case '7d':
        cutoffTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case '30d':
        cutoffTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      default:
        cutoffTime = new Date(0); // Show all
    }
    
    filtered = filtered.filter(log => new Date(log.timestamp) >= cutoffTime);
    
    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(log => 
        log.title.toLowerCase().includes(query) ||
        log.description.toLowerCase().includes(query) ||
        log.metadata?.repositoryName?.toLowerCase().includes(query) ||
        log.metadata?.taskTitle?.toLowerCase().includes(query)
      );
    }
    
    return filtered;
  }, [logs, searchQuery, selectedType, selectedDateRange, repository.id, repository.name]);
  
  const handleExportLogs = () => {
    const exportData = activityLogger.exportLogs({
      format: 'csv',
      includeMetadata: true,
      filters: {
        type: selectedType === 'all' ? undefined : selectedType,
        searchQuery: searchQuery.trim() || undefined,
        repositoryId: repository.id
      }
    });
    
    const blob = new Blob([exportData], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${repository.name}_activity_logs_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    window.URL.revokeObjectURL(url);
  };

  const handleClearLogs = () => {
    if (confirm('Are you sure you want to clear all activity logs? This action cannot be undone.')) {
      activityLogger.clearLogs();
    }
  };
  
  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Activity Logs
          </h2>
          {stats && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <BarChart3 className="h-4 w-4" />
              <span>{stats.total} total</span>
              <span>•</span>
              <span>{stats.recentActivity} recent</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleExportLogs}
            disabled={filteredLogs.length === 0}
          >
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleClearLogs}
            disabled={logs.length === 0}
            className="text-red-600 hover:text-red-700 hover:bg-red-50"
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Clear
          </Button>
        </div>
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
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search activity logs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
            
            {/* Activity Type Filter */}
            <Select value={selectedType} onValueChange={(value) => setSelectedType(value as ActivityType)}>
              <SelectTrigger>
                <SelectValue placeholder="Filter by type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Activities</SelectItem>
                <SelectItem value="connection">Connection</SelectItem>
                <SelectItem value="task">Tasks</SelectItem>
                <SelectItem value="github">GitHub</SelectItem>
                <SelectItem value="navigation">Navigation</SelectItem>
              </SelectContent>
            </Select>
            
            {/* Date Range Filter */}
            <Select value={selectedDateRange} onValueChange={setSelectedDateRange}>
              <SelectTrigger>
                <SelectValue placeholder="Select time range" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1h">Last Hour</SelectItem>
                <SelectItem value="24h">Last 24 Hours</SelectItem>
                <SelectItem value="7d">Last 7 Days</SelectItem>
                <SelectItem value="30d">Last 30 Days</SelectItem>
                <SelectItem value="all">All Time</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>
      
      {/* Activity Feed */}
      <div className="space-y-3">
        {filteredLogs.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Activity className="h-12 w-12 mb-4 opacity-50" />
              <h3 className="text-lg font-semibold mb-2">No activity logs found</h3>
              <p className="text-sm text-muted-foreground text-center">
                {searchQuery || selectedType !== 'all' || selectedDateRange !== '24h'
                  ? 'Try adjusting your filters to see more results'
                  : 'Activity logs will appear here when tasks are executed and repository interactions occur'
                }
              </p>
            </CardContent>
          </Card>
        ) : (
          <>
            <div className="flex items-center justify-between text-sm text-muted-foreground mb-2">
              <span>Showing {filteredLogs.length} log{filteredLogs.length === 1 ? '' : 's'}</span>
            </div>
            
            {filteredLogs.map((log, index) => (
              <Card key={log.id} className="overflow-hidden">
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    {/* Activity Icon */}
                    <div className="mt-1">
                      {getActivityIcon(log.type, log.level)}
                    </div>
                    
                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="font-medium text-sm">{log.title}</h3>
                        <Badge 
                          variant="secondary" 
                          className={`text-xs ${getActivityBadgeColor(log.type)}`}
                        >
                          {log.type}
                        </Badge>
                      </div>
                      
                      <p className="text-sm text-muted-foreground mb-2">
                        {log.description}
                      </p>
                      
                      {/* Metadata */}
                      {log.metadata && Object.keys(log.metadata).length > 0 && (
                        <div className="flex flex-wrap gap-2 mb-2">
                          {log.metadata.taskTitle && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-blue-50 text-blue-700 px-2 py-1 rounded">
                              <CheckCircle className="h-3 w-3" />
                              {log.metadata.taskTitle}
                            </div>
                          )}
                          {log.metadata.branchName && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                              <GitBranch className="h-3 w-3" />
                              {log.metadata.branchName}
                            </div>
                          )}
                          {(log.metadata.prUrl || log.metadata.githubUrl) && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 px-2 text-xs"
                              asChild
                            >
                              <a href={log.metadata.prUrl || log.metadata.githubUrl} target="_blank" rel="noopener noreferrer">
                                <GitPullRequest className="h-3 w-3 mr-1" />
                                {log.metadata.prNumber ? `PR #${log.metadata.prNumber}` : 'View'}
                                <ExternalLink className="h-3 w-3 ml-1" />
                              </a>
                            </Button>
                          )}
                          {log.metadata.issueNumber && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-green-50 text-green-700 px-2 py-1 rounded">
                              <AlertCircle className="h-3 w-3" />
                              Issue #{log.metadata.issueNumber}
                            </div>
                          )}
                          {log.metadata.duration && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                              <Clock className="h-3 w-3" />
                              {Math.round(log.metadata.duration / 1000)}s
                            </div>
                          )}
                          {log.metadata.rateLimitRemaining !== undefined && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-yellow-50 text-yellow-700 px-2 py-1 rounded">
                              <AlertCircle className="h-3 w-3" />
                              Rate limit: {log.metadata.rateLimitRemaining}
                            </div>
                          )}
                          {log.metadata.errorMessage && (
                            <div className="flex items-center gap-1 text-xs text-red-700 bg-red-50 px-2 py-1 rounded max-w-xs truncate">
                              <XCircle className="h-3 w-3 shrink-0" />
                              <span title={log.metadata.errorMessage}>
                                {log.metadata.errorMessage.length > 30 
                                  ? `${log.metadata.errorMessage.substring(0, 30)}...` 
                                  : log.metadata.errorMessage
                                }
                              </span>
                            </div>
                          )}
                        </div>
                      )}
                      
                      {/* Timestamp */}
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <Calendar className="h-3 w-3" />
                        <time dateTime={log.timestamp}>
                          {formatRelativeTime(log.timestamp)} • {new Date(log.timestamp).toLocaleString()}
                        </time>
                      </div>
                    </div>
                  </div>
                </CardContent>
                
                {/* Separator for all items except the last */}
                {index < filteredLogs.length - 1 && <Separator />}
              </Card>
            ))}
          </>
        )}
      </div>
    </div>
  );
}