'use client';

import { useState, useEffect } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../components/ui/tabs';
import { Button } from '../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card';
import { Badge } from '../../components/ui/badge';
import { ExternalLink, Activity, BarChart3, AlertCircle, CheckCircle } from 'lucide-react';
import { TaskTab } from './tabs/TaskTab';
import { LogsTab } from './tabs/LogsTab';
import { DashboardTab } from './tabs/DashboardTab';
import { ErrorBoundary } from '../../components/ErrorBoundary';

interface WorkspacePanelProps {
  repository: Repository;
  onClose: () => void;
}


export function WorkspacePanel({ repository, onClose }: WorkspacePanelProps) {
  const [activeTab, setActiveTab] = useState<string>('tasks');

  // Clear all caches when repository changes and restore tab state
  useEffect(() => {
    // Aggressive cache clearing for new repository
    try {
      // Clear all possible cache keys
      Object.keys(localStorage).forEach(key => {
        if (key.includes('tasks') || key.includes('cache')) {
          localStorage.removeItem(key);
        }
      });
      Object.keys(sessionStorage).forEach(key => {
        if (key.includes('tasks') || key.includes('cache')) {
          sessionStorage.removeItem(key);
        }
      });
    } catch (error) {
      // Ignore storage errors
      console.warn('Failed to clear cache:', error);
    }

    // Restore tab state for this repository
    const savedTab = localStorage.getItem(`workspace-tab-${repository.id}`);
    if (savedTab && ['tasks', 'logs', 'dashboard'].includes(savedTab)) {
      setActiveTab(savedTab);
    }
  }, [repository.id, repository.name]);

  const handleTabChange = (value: string) => {
    setActiveTab(value);
    localStorage.setItem(`workspace-tab-${repository.id}`, value);
  };

  // Show connection message for non-connected repositories
  if (!repository.is_connected) {
    return (
      <div className="fixed inset-0 bg-background z-50 flex flex-col safe-area-inset">
        {/* Header */}
        <div className="border-b bg-card">
          <div className="container mx-auto max-w-7xl px-4 py-4 px-safe-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <h1 className="text-xl md:text-2xl font-bold">{repository.name}</h1>
                  <Badge variant={repository.private ? "secondary" : "outline"}>
                    {repository.private ? 'Private' : 'Public'}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">{repository.description}</p>
              </div>
              <Button variant="outline" onClick={onClose} className="shrink-0">
                ← Back
              </Button>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 container mx-auto max-w-7xl px-4 py-6 px-safe-4">
          <div className="flex flex-col items-center justify-center h-full text-center space-y-6">
            <AlertCircle className="h-16 w-16 text-muted-foreground" />
            <div className="space-y-2">
              <h2 className="text-2xl font-bold">Repository Not Connected</h2>
              <p className="text-muted-foreground max-w-md">
                This repository needs to be connected before you can access the workspace features.
                Please connect the repository first to view tasks, logs, and dashboard.
              </p>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={onClose}>
                ← Back to Repositories
              </Button>
              <Button variant="outline" asChild>
                <a href={repository.html_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4 mr-2" />
                  View on GitHub
                </a>
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // For connected repositories, show the 3-tab interface
  return (
    <div className="fixed inset-0 bg-background z-50 flex flex-col safe-area-inset">
      {/* Header */}
      <div className="border-b bg-card">
        <div className="container mx-auto max-w-7xl px-4 py-4 px-safe-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <h1 className="text-xl md:text-2xl font-bold">{repository.name}</h1>
                <Badge variant={repository.private ? "secondary" : "outline"}>
                  {repository.private ? 'Private' : 'Public'}
                </Badge>
                <Badge variant="default" className="bg-green-100 text-green-800">
                  Connected
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">{repository.description}</p>
            </div>
            <div className="flex gap-2 items-center">
              <Button variant="outline" size="sm" asChild>
                <a href={repository.html_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4 mr-2" />
                  View on GitHub
                </a>
              </Button>
              <Button variant="outline" onClick={onClose} className="shrink-0">
                ← Back
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        <div className="container mx-auto max-w-7xl px-4 py-6 px-safe-4 h-full">
          <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full h-full flex flex-col">
          <TabsList className="grid w-full grid-cols-3 mb-6">
            <TabsTrigger value="tasks" className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4" />
              <span className="hidden sm:inline">Tasks</span>
            </TabsTrigger>
            <TabsTrigger value="logs" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              <span className="hidden sm:inline">Logs</span>
            </TabsTrigger>
            <TabsTrigger value="dashboard" className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              <span className="hidden sm:inline">Dashboard</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="tasks" className="space-y-4 flex-1 overflow-y-auto min-h-0">
            <ErrorBoundary 
              level="component" 
              showDetails={process.env.NODE_ENV === 'development'}
              onError={(error, errorInfo) => {
                console.error('TaskTab Error:', error, errorInfo);
                // Log to activity logger if available
              }}
              fallback={(error, resetError) => (
                <Card className="border-red-200 bg-red-50">
                  <CardHeader>
                    <CardTitle className="text-red-900 flex items-center gap-2">
                      <AlertCircle className="h-5 w-5" />
                      Tasks Tab Error
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-red-700 mb-4">
                      The tasks tab encountered an error and couldn&apos;t load properly.
                    </p>
                    <div className="flex gap-2">
                      <Button variant="outline" onClick={resetError}>
                        Try Again
                      </Button>
                      <Button 
                        variant="outline" 
                        onClick={() => window.location.reload()}
                      >
                        Refresh Page
                      </Button>
                    </div>
                    {process.env.NODE_ENV === 'development' && (
                      <details className="mt-4 text-xs">
                        <summary className="cursor-pointer font-medium">Technical Details</summary>
                        <pre className="mt-2 p-2 bg-red-100 rounded overflow-auto">
                          {error.message}
                        </pre>
                      </details>
                    )}
                  </CardContent>
                </Card>
              )}
            >
              <TaskTab repository={repository} />
            </ErrorBoundary>
          </TabsContent>

          <TabsContent value="logs" className="space-y-4 flex-1 overflow-y-auto min-h-0">
            <ErrorBoundary 
              level="component" 
              showDetails={process.env.NODE_ENV === 'development'}
              onError={(error, errorInfo) => {
                console.error('LogsTab Error:', error, errorInfo);
              }}
              fallback={(error, resetError) => (
                <Card className="border-red-200 bg-red-50">
                  <CardHeader>
                    <CardTitle className="text-red-900 flex items-center gap-2">
                      <AlertCircle className="h-5 w-5" />
                      Logs Tab Error
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-red-700 mb-4">
                      The activity logs tab encountered an error and couldn&apos;t load properly.
                    </p>
                    <div className="flex gap-2">
                      <Button variant="outline" onClick={resetError}>
                        Try Again
                      </Button>
                      <Button 
                        variant="outline" 
                        onClick={() => window.location.reload()}
                      >
                        Refresh Page
                      </Button>
                    </div>
                    {process.env.NODE_ENV === 'development' && (
                      <details className="mt-4 text-xs">
                        <summary className="cursor-pointer font-medium">Technical Details</summary>
                        <pre className="mt-2 p-2 bg-red-100 rounded overflow-auto">
                          {error.message}
                        </pre>
                      </details>
                    )}
                  </CardContent>
                </Card>
              )}
            >
              <LogsTab repository={repository} />
            </ErrorBoundary>
          </TabsContent>

          <TabsContent value="dashboard" className="space-y-4 flex-1 overflow-y-auto min-h-0">
            <ErrorBoundary 
              level="component" 
              showDetails={process.env.NODE_ENV === 'development'}
              onError={(error, errorInfo) => {
                console.error('DashboardTab Error:', error, errorInfo);
              }}
              fallback={(error, resetError) => (
                <Card className="border-red-200 bg-red-50">
                  <CardHeader>
                    <CardTitle className="text-red-900 flex items-center gap-2">
                      <AlertCircle className="h-5 w-5" />
                      Dashboard Tab Error
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-red-700 mb-4">
                      The dashboard tab encountered an error and couldn&apos;t load properly.
                    </p>
                    <div className="flex gap-2">
                      <Button variant="outline" onClick={resetError}>
                        Try Again
                      </Button>
                      <Button 
                        variant="outline" 
                        onClick={() => window.location.reload()}
                      >
                        Refresh Page
                      </Button>
                    </div>
                    {process.env.NODE_ENV === 'development' && (
                      <details className="mt-4 text-xs">
                        <summary className="cursor-pointer font-medium">Technical Details</summary>
                        <pre className="mt-2 p-2 bg-red-100 rounded overflow-auto">
                          {error.message}
                        </pre>
                      </details>
                    )}
                  </CardContent>
                </Card>
              )}
            >
              <DashboardTab repository={repository} />
            </ErrorBoundary>
          </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  );
}