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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '../../../components/ui/dialog';
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
  BarChart3,
  Eye,
  Info
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
  const [selectedLog, setSelectedLog] = useState<ActivityLog | null>(null);
  
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

  const handleLogClick = (log: ActivityLog) => {
    setSelectedLog(log);
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
              <Card 
                key={log.id} 
                className="overflow-hidden cursor-pointer hover:bg-muted/50 transition-colors group"
                onClick={() => handleLogClick(log)}
              >
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    {/* Activity Icon */}
                    <div className="mt-1">
                      {getActivityIcon(log.type, log.level)}
                    </div>
                    
                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between mb-1">
                        <div className="flex items-center gap-2">
                          <h3 className="font-medium text-sm">{log.title}</h3>
                          <Badge 
                            variant="secondary" 
                            className={`text-xs ${getActivityBadgeColor(log.type)}`}
                          >
                            {log.type}
                          </Badge>
                        </div>
                        <Button 
                          variant="ghost" 
                          size="sm" 
                          className="opacity-0 group-hover:opacity-100 transition-opacity"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleLogClick(log);
                          }}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                      </div>
                      
                      <p className="text-sm text-muted-foreground mb-2 line-clamp-2">
                        {log.description.length > 100 
                          ? `${log.description.substring(0, 100)}...` 
                          : log.description
                        }
                      </p>
                      
                      {/* Click hint */}
                      <div className="text-xs text-muted-foreground/70 mb-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        Click to view full details
                      </div>
                      
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

      {/* Log Detail Modal */}
      <Dialog open={!!selectedLog} onOpenChange={() => setSelectedLog(null)}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          {selectedLog && (
            <>
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <div className="flex items-center gap-2">
                    {getActivityIcon(selectedLog.type, selectedLog.level)}
                    <span>{selectedLog.title}</span>
                    <Badge 
                      variant="secondary" 
                      className={`text-xs ${getActivityBadgeColor(selectedLog.type)}`}
                    >
                      {selectedLog.type}
                    </Badge>
                  </div>
                </DialogTitle>
                <DialogDescription>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Calendar className="h-4 w-4" />
                    <time dateTime={selectedLog.timestamp}>
                      {new Date(selectedLog.timestamp).toLocaleString()}
                    </time>
                    <span>•</span>
                    <span>{formatRelativeTime(selectedLog.timestamp)}</span>
                  </div>
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4">
                {/* Main Description */}
                <div>
                  <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                    <Info className="h-4 w-4" />
                    Description
                  </h4>
                  <p className="text-sm text-muted-foreground bg-muted p-3 rounded-md">
                    {selectedLog.description}
                  </p>
                </div>

                {/* Activity Level */}
                <div>
                  <h4 className="text-sm font-medium mb-2">Level</h4>
                  <Badge 
                    variant={selectedLog.level === 'error' ? 'destructive' : selectedLog.level === 'warning' ? 'secondary' : 'default'}
                    className="capitalize"
                  >
                    {selectedLog.level}
                  </Badge>
                </div>

                {/* Metadata Details */}
                {selectedLog.metadata && Object.keys(selectedLog.metadata).length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-3">Metadata</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {selectedLog.metadata.repositoryName && (
                        <div className="bg-muted p-3 rounded-md">
                          <div className="text-xs font-medium text-muted-foreground mb-1">Repository</div>
                          <div className="text-sm">{selectedLog.metadata.repositoryName}</div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.taskTitle && (
                        <div className="bg-blue-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-blue-600 mb-1">Task</div>
                          <div className="text-sm text-blue-800">{selectedLog.metadata.taskTitle}</div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.branchName && (
                        <div className="bg-purple-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-purple-600 mb-1">Branch</div>
                          <div className="text-sm text-purple-800 flex items-center gap-1">
                            <GitBranch className="h-3 w-3" />
                            {selectedLog.metadata.branchName}
                          </div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.duration && (
                        <div className="bg-green-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-green-600 mb-1">Duration</div>
                          <div className="text-sm text-green-800 flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {Math.round(selectedLog.metadata.duration / 1000)}s ({selectedLog.metadata.duration}ms)
                          </div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.issueNumber && (
                        <div className="bg-orange-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-orange-600 mb-1">GitHub Issue</div>
                          <div className="text-sm text-orange-800 flex items-center gap-1">
                            <AlertCircle className="h-3 w-3" />
                            Issue #{selectedLog.metadata.issueNumber}
                          </div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.prNumber && (
                        <div className="bg-cyan-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-cyan-600 mb-1">Pull Request</div>
                          <div className="text-sm text-cyan-800 flex items-center gap-1">
                            <GitPullRequest className="h-3 w-3" />
                            PR #{selectedLog.metadata.prNumber}
                          </div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.rateLimitRemaining !== undefined && (
                        <div className="bg-yellow-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-yellow-600 mb-1">Rate Limit</div>
                          <div className="text-sm text-yellow-800">
                            {selectedLog.metadata.rateLimitRemaining} requests remaining
                          </div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.userAgent && (
                        <div className="bg-gray-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-gray-600 mb-1">User Agent</div>
                          <div className="text-xs text-gray-700 font-mono truncate" title={selectedLog.metadata.userAgent}>
                            {selectedLog.metadata.userAgent}
                          </div>
                        </div>
                      )}
                      
                      {selectedLog.metadata.requestId && (
                        <div className="bg-gray-50 p-3 rounded-md">
                          <div className="text-xs font-medium text-gray-600 mb-1">Request ID</div>
                          <div className="text-xs text-gray-700 font-mono">
                            {selectedLog.metadata.requestId}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )}
                
                {/* Error Message */}
                {selectedLog.metadata?.errorMessage && (
                  <div>
                    <h4 className="text-sm font-medium mb-2 text-red-600 flex items-center gap-2">
                      <XCircle className="h-4 w-4" />
                      Error Details
                    </h4>
                    <div className="bg-red-50 border border-red-200 p-3 rounded-md">
                      <pre className="text-sm text-red-800 whitespace-pre-wrap font-mono">
                        {selectedLog.metadata.errorMessage}
                      </pre>
                    </div>
                  </div>
                )}
                
                {/* External Links */}
                {(selectedLog.metadata?.prUrl || selectedLog.metadata?.githubUrl) && (
                  <div className="flex gap-2">
                    {selectedLog.metadata.prUrl && (
                      <Button variant="outline" asChild>
                        <a href={selectedLog.metadata.prUrl} target="_blank" rel="noopener noreferrer">
                          <GitPullRequest className="h-4 w-4 mr-2" />
                          View Pull Request
                          <ExternalLink className="h-4 w-4 ml-2" />
                        </a>
                      </Button>
                    )}
                    {selectedLog.metadata.githubUrl && (
                      <Button variant="outline" asChild>
                        <a href={selectedLog.metadata.githubUrl} target="_blank" rel="noopener noreferrer">
                          <ExternalLink className="h-4 w-4 mr-2" />
                          View on GitHub
                        </a>
                      </Button>
                    )}
                  </div>
                )}
                
                <div className="flex justify-end">
                  <Button variant="outline" onClick={() => setSelectedLog(null)}>
                    Close
                  </Button>
                </div>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}