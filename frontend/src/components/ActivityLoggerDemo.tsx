/**
 * Activity Logger Demo Component
 * Demonstrates the activity logging system functionality
 * This component can be temporarily added to test the logging system
 */

'use client';

import { useEffect, useState } from 'react';
import { ActivityLogger } from '../services/ActivityLogger';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Separator } from './ui/separator';

export function ActivityLoggerDemo() {
  const [activityLogger] = useState(() => ActivityLogger.getInstance());
  const [logs, setLogs] = useState<any[]>([]);
  const [stats, setStats] = useState<any>(null);

  useEffect(() => {
    // Subscribe to activity updates
    const unsubscribe = activityLogger.subscribe((updatedLogs) => {
      setLogs(updatedLogs);
      setStats(activityLogger.getStatistics());
    });

    // Initial load
    setLogs(activityLogger.getLogs());
    setStats(activityLogger.getStatistics());

    return unsubscribe;
  }, [activityLogger]);

  const generateSampleActivities = () => {
    // Simulate repository connection
    activityLogger.logRepositoryConnected(12345, 'workflow', '/Users/dev/workflow');
    
    setTimeout(() => {
      // Simulate GitHub sync
      activityLogger.logGitHubSync('workflow', 'started');
      
      setTimeout(() => {
        // Simulate API calls
        activityLogger.logGitHubApiCall('/repos/workflow/issues', 'GET', 4500);
        activityLogger.logGitHubApiCall('/repos/workflow/pulls', 'GET', 4499);
        
        setTimeout(() => {
          // Complete sync
          activityLogger.logGitHubSync('workflow', 'completed', { duration: 2500, apiCallCount: 2 });
          
          // Simulate task creation
          activityLogger.logTaskCreated(
            'task-demo-1',
            'Implement activity logging system', 
            12345,
            'workflow',
            { branchName: 'feature/activity-logging', githubUrl: 'https://github.com/workflow/issues/32' }
          );
          
          setTimeout(() => {
            // Start task
            activityLogger.logTaskStarted('task-demo-1', 'Implement activity logging system');
            
            setTimeout(() => {
              // Complete task
              activityLogger.logTaskCompleted('task-demo-1', 'Implement activity logging system', 180000);
              
              // Simulate navigation
              activityLogger.logTabSwitch('tasks', 'logs', 'workflow');
              
              // Simulate rate limit warning
              setTimeout(() => {
                activityLogger.logGitHubRateLimit(150, new Date(Date.now() + 3600000).toLocaleTimeString());
              }, 1000);
              
            }, 500);
          }, 300);
        }, 400);
      }, 200);
    }, 100);
  };

  const clearAllLogs = () => {
    activityLogger.clearLogs();
  };

  const exportLogs = () => {
    const csvData = activityLogger.exportLogs({ format: 'csv', includeMetadata: true });
    const blob = new Blob([csvData], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `activity_logs_demo_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Activity Logger Demo</span>
          <div className="flex gap-2">
            <Button onClick={generateSampleActivities} size="sm">
              Generate Sample Activities
            </Button>
            <Button onClick={exportLogs} variant="outline" size="sm" disabled={logs.length === 0}>
              Export CSV
            </Button>
            <Button onClick={clearAllLogs} variant="outline" size="sm" disabled={logs.length === 0}>
              Clear All
            </Button>
          </div>
        </CardTitle>
      </CardHeader>
      
      <CardContent className="space-y-6">
        {/* Statistics */}
        {stats && (
          <div>
            <h3 className="font-medium mb-3">Statistics</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
              <div className="text-center p-3 border rounded-lg">
                <div className="text-2xl font-bold text-blue-600">{stats.total}</div>
                <div className="text-sm text-muted-foreground">Total Logs</div>
              </div>
              <div className="text-center p-3 border rounded-lg">
                <div className="text-2xl font-bold text-green-600">{stats.recentActivity}</div>
                <div className="text-sm text-muted-foreground">Last 24h</div>
              </div>
              <div className="text-center p-3 border rounded-lg">
                <div className="text-2xl font-bold text-purple-600">{stats.byLevel.success}</div>
                <div className="text-sm text-muted-foreground">Success</div>
              </div>
              <div className="text-center p-3 border rounded-lg">
                <div className="text-2xl font-bold text-red-600">{stats.byLevel.error}</div>
                <div className="text-sm text-muted-foreground">Errors</div>
              </div>
            </div>
            
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline">Connection: {stats.byType.connection}</Badge>
              <Badge variant="outline">Tasks: {stats.byType.task}</Badge>
              <Badge variant="outline">GitHub: {stats.byType.github}</Badge>
              <Badge variant="outline">Navigation: {stats.byType.navigation}</Badge>
            </div>
          </div>
        )}

        <Separator />

        {/* Recent Activity */}
        <div>
          <h3 className="font-medium mb-3">Recent Activity ({logs.length} logs)</h3>
          <div className="max-h-96 overflow-y-auto space-y-2">
            {logs.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No activity logs yet. Click "Generate Sample Activities" to see the system in action.
              </div>
            ) : (
              logs.slice(0, 20).map((log) => (
                <div key={log.id} className="border rounded-lg p-3">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <Badge
                          className={
                            log.type === 'connection' ? 'bg-green-100 text-green-800' :
                            log.type === 'task' ? 'bg-blue-100 text-blue-800' :
                            log.type === 'github' ? 'bg-purple-100 text-purple-800' :
                            'bg-gray-100 text-gray-800'
                          }
                        >
                          {log.type}
                        </Badge>
                        <Badge
                          className={
                            log.level === 'success' ? 'bg-green-50 text-green-700' :
                            log.level === 'error' ? 'bg-red-50 text-red-700' :
                            log.level === 'warning' ? 'bg-yellow-50 text-yellow-700' :
                            'bg-blue-50 text-blue-700'
                          }
                        >
                          {log.level}
                        </Badge>
                        <span className="font-medium text-sm truncate">{log.title}</span>
                      </div>
                      <p className="text-sm text-muted-foreground mb-2">{log.description}</p>
                      
                      {/* Metadata */}
                      {log.metadata && Object.keys(log.metadata).length > 0 && (
                        <div className="flex flex-wrap gap-1 mb-2">
                          {log.metadata.taskTitle && (
                            <Badge variant="outline" className="text-xs">{log.metadata.taskTitle}</Badge>
                          )}
                          {log.metadata.branchName && (
                            <Badge variant="outline" className="text-xs">🌿 {log.metadata.branchName}</Badge>
                          )}
                          {log.metadata.duration && (
                            <Badge variant="outline" className="text-xs">
                              ⏱️ {Math.round(log.metadata.duration / 1000)}s
                            </Badge>
                          )}
                          {log.metadata.rateLimitRemaining !== undefined && (
                            <Badge variant="outline" className="text-xs">
                              📊 Rate limit: {log.metadata.rateLimitRemaining}
                            </Badge>
                          )}
                        </div>
                      )}
                      
                      <div className="text-xs text-muted-foreground">
                        {new Date(log.timestamp).toLocaleString()}
                      </div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}