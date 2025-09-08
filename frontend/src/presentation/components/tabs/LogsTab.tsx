'use client';

import { useState, useMemo } from 'react';
import { Repository } from '../../../domain/entities/Repository';
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
  ExternalLink
} from 'lucide-react';

interface LogsTabProps {
  repository: Repository;
}

// Activity log types for filtering
type ActivityType = 'connection' | 'task' | 'github' | 'navigation' | 'all';
type ActivityLevel = 'info' | 'success' | 'warning' | 'error';

interface ActivityLog {
  id: string;
  timestamp: string;
  type: ActivityType;
  level: ActivityLevel;
  title: string;
  description: string;
  metadata?: {
    taskId?: string;
    branchName?: string;
    prUrl?: string;
    issueNumber?: number;
    duration?: number;
    userId?: string;
  };
}

// Mock activity logs data - in a real app, this would come from a service
const generateMockLogs = (repository: Repository): ActivityLog[] => {
  const baseTime = new Date();
  return [
    {
      id: '1',
      timestamp: new Date(baseTime.getTime() - 5 * 60 * 1000).toISOString(),
      type: 'github',
      level: 'success',
      title: 'Repository synchronized',
      description: `Successfully fetched latest issues and pull requests from ${repository.name}`,
      metadata: {
        duration: 1200
      }
    },
    {
      id: '2',
      timestamp: new Date(baseTime.getTime() - 15 * 60 * 1000).toISOString(),
      type: 'task',
      level: 'info',
      title: 'Task created',
      description: 'New task "Add dark mode support" created and assigned to repository',
      metadata: {
        taskId: 'task-1',
        branchName: 'feature/dark-mode'
      }
    },
    {
      id: '3',
      timestamp: new Date(baseTime.getTime() - 30 * 60 * 1000).toISOString(),
      type: 'github',
      level: 'success',
      title: 'Pull request merged',
      description: 'PR #42: "Fix responsive design issues" was successfully merged',
      metadata: {
        prUrl: `${repository.html_url}/pull/42`
      }
    },
    {
      id: '4',
      timestamp: new Date(baseTime.getTime() - 45 * 60 * 1000).toISOString(),
      type: 'connection',
      level: 'success',
      title: 'Repository connected',
      description: `Successfully connected to ${repository.name} repository`,
      metadata: {
        duration: 800
      }
    },
    {
      id: '5',
      timestamp: new Date(baseTime.getTime() - 60 * 60 * 1000).toISOString(),
      type: 'navigation',
      level: 'info',
      title: 'Workspace opened',
      description: `Accessed workspace for ${repository.name}`,
      metadata: {}
    },
    {
      id: '6',
      timestamp: new Date(baseTime.getTime() - 75 * 60 * 1000).toISOString(),
      type: 'github',
      level: 'warning',
      title: 'API rate limit warning',
      description: 'GitHub API rate limit at 80%, consider reducing request frequency',
      metadata: {}
    },
    {
      id: '7',
      timestamp: new Date(baseTime.getTime() - 120 * 60 * 1000).toISOString(),
      type: 'task',
      level: 'success',
      title: 'Task completed',
      description: 'Task "Implement authentication system" completed successfully',
      metadata: {
        taskId: 'task-7',
        branchName: 'feature/auth-system',
        prUrl: `${repository.html_url}/pull/41`,
        duration: 3600
      }
    }
  ];
};

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
  const [selectedType, setSelectedType] = useState<ActivityType>('all');
  const [selectedDateRange, setSelectedDateRange] = useState<string>('24h');
  
  // Generate mock logs
  const mockLogs = generateMockLogs(repository);
  
  // Filter logs based on search and filters
  const filteredLogs = useMemo(() => {
    let filtered = mockLogs;
    
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
        log.description.toLowerCase().includes(query)
      );
    }
    
    return filtered.sort((a, b) => 
      new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    );
  }, [mockLogs, searchQuery, selectedType, selectedDateRange]);
  
  const handleExportLogs = () => {
    const csvContent = [
      ['Timestamp', 'Type', 'Level', 'Title', 'Description'].join(','),
      ...filteredLogs.map(log => [
        log.timestamp,
        log.type,
        log.level,
        `"${log.title}"`,
        `"${log.description}"`
      ].join(','))
    ].join('\n');
    
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${repository.name}_activity_logs_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    window.URL.revokeObjectURL(url);
  };
  
  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Activity className="h-5 w-5" />
          Activity Logs
        </h2>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={handleExportLogs}
          disabled={filteredLogs.length === 0}
        >
          <Download className="h-4 w-4 mr-2" />
          Export
        </Button>
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
                          {log.metadata.branchName && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                              <GitBranch className="h-3 w-3" />
                              {log.metadata.branchName}
                            </div>
                          )}
                          {log.metadata.prUrl && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 px-2 text-xs"
                              asChild
                            >
                              <a href={log.metadata.prUrl} target="_blank" rel="noopener noreferrer">
                                <GitPullRequest className="h-3 w-3 mr-1" />
                                View PR
                                <ExternalLink className="h-3 w-3 ml-1" />
                              </a>
                            </Button>
                          )}
                          {log.metadata.duration && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                              <Clock className="h-3 w-3" />
                              {Math.round(log.metadata.duration / 1000)}s
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