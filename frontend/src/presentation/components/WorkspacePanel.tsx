'use client';

import { useState, useEffect } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../components/ui/tabs';
import { Button } from '../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card';
import { Badge } from '../../components/ui/badge';
import { ExternalLink, Activity, AlertCircle, CheckCircle } from 'lucide-react';
import { TaskTab } from './tabs/TaskTab';
import { ErrorBoundary } from '../../components/ErrorBoundary';

interface WorkspacePanelProps {
  repository: Repository;
  onClose: () => void;
}


export function WorkspacePanel({ repository, onClose }: WorkspacePanelProps) {
  const [activeTab, setActiveTab] = useState<string>('tasks');

  // Clear all caches when repository changes and restore tab state
  useEffect(() => {
    // Complete cache clearing function
    const clearAllCaches = () => {
      try {
        console.log('🧹 Starting complete cache cleanup for repository:', repository.name);
        
        // Clear localStorage completely
        Object.keys(localStorage).forEach(key => {
          if (key.includes('tasks') || key.includes('cache') || key.includes('github') || key.includes('repo') || key.includes('issues') || key.includes('prs')) {
            localStorage.removeItem(key);
            console.log('Cleared localStorage key:', key);
          }
        });
        
        // Clear sessionStorage completely  
        Object.keys(sessionStorage).forEach(key => {
          if (key.includes('tasks') || key.includes('cache') || key.includes('github') || key.includes('repo') || key.includes('issues') || key.includes('prs')) {
            sessionStorage.removeItem(key);
            console.log('Cleared sessionStorage key:', key);
          }
        });
        
        // Clear IndexedDB if available
        if ('indexedDB' in window) {
          try {
            indexedDB.deleteDatabase('cache-db');
            indexedDB.deleteDatabase('tasks-db');
            console.log('Cleared IndexedDB databases');
          } catch (e) {
            // Ignore IndexedDB errors
          }
        }
        
        // Clear any service worker caches if available
        if ('serviceWorker' in navigator && 'caches' in window) {
          caches.keys().then(cacheNames => {
            const deletePromises = cacheNames
              .filter(cacheName => 
                cacheName.includes('api') || 
                cacheName.includes('tasks') || 
                cacheName.includes('github') ||
                cacheName.includes('next')
              )
              .map(cacheName => {
                console.log('Deleting cache:', cacheName);
                return caches.delete(cacheName);
              });
            return Promise.all(deletePromises);
          }).then(() => {
            console.log('Service Worker caches cleared');
          }).catch(() => {
            // Ignore cache errors
          });
        }
        
        // Force clear any React Query cache if available
        if (typeof window !== 'undefined' && (window as any).queryClient) {
          (window as any).queryClient.clear();
          console.log('React Query cache cleared');
        }
        
        console.log('🧽 Complete cache cleanup completed for repository:', repository.name);
      } catch (error) {
        console.warn('Cache clearing failed:', error);
      }
    };
    
    // Execute cache clearing immediately
    clearAllCaches();
    
    // Also clear cache when component unmounts
    return () => {
      clearAllCaches();
    };

    // Restore tab state for this repository
    const savedTab = localStorage.getItem(`workspace-tab-${repository.id}`);
    if (savedTab && ['tasks', 'issues', 'prs'].includes(savedTab)) {
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
      <div className="fixed inset-0 bg-background z-50 flex flex-col">
        {/* Header */}
        <div className="border-b bg-card">
          <div className="container mx-auto max-w-7xl px-4 py-4">
            <div className="flex items-center justify-between">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <h1 className="text-lg md:text-2xl font-bold truncate">{repository.name}</h1>
                  <Badge variant={repository.private ? "secondary" : "outline"} className="text-xs md:text-sm">
                    {repository.private ? 'Private' : 'Public'}
                  </Badge>
                  <Badge variant="destructive" className="text-xs md:text-sm">
                    Not Connected
                  </Badge>
                </div>
                {repository.description && (
                  <p className="text-sm text-muted-foreground truncate">{repository.description}</p>
                )}
              </div>
              <Button variant="outline" onClick={onClose} className="shrink-0">
                ← Back
              </Button>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 container mx-auto max-w-7xl px-4 py-6">
          <div className="flex flex-col items-center justify-center h-full text-center space-y-6">
            <AlertCircle className="h-16 w-16 text-muted-foreground" />
            <div className="space-y-2">
              <h2 className="text-2xl font-bold">Repository Not Connected</h2>
              <p className="text-muted-foreground max-w-md">
                This repository needs to be connected before you can access the workspace features.
                Please connect the repository first to view tasks, logs, and dashboard.
              </p>
            </div>
            <div className="flex flex-col sm:flex-row gap-2">
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

  // For connected repositories, show the responsive interface
  return (
    <div className="fixed inset-0 bg-background z-50 flex flex-col">
      {/* Responsive Header */}
      <div className="border-b bg-card">
        <div className="container mx-auto max-w-7xl px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <h1 className="text-lg md:text-2xl font-bold truncate">{repository.name}</h1>
                <Badge variant={repository.private ? "secondary" : "outline"} className="text-xs md:text-sm">
                  {repository.private ? 'Private' : 'Public'}
                </Badge>
                <Badge variant="default" className="bg-green-100 text-green-800 text-xs md:text-sm">
                  Connected
                </Badge>
              </div>
              {repository.description && (
                <p className="text-sm text-muted-foreground truncate">{repository.description}</p>
              )}
            </div>
            <div className="flex gap-2 items-center">
              <Button variant="outline" size="sm" asChild>
                <a href={repository.html_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2">
                  <ExternalLink className="h-4 w-4" />
                  <span className="hidden sm:inline">View on GitHub</span>
                  <span className="sm:hidden">GitHub</span>
                </a>
              </Button>
              <Button variant="outline" onClick={onClose} className="shrink-0">
                ← Back
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Content with responsive tabs */}
      <div className="flex-1 overflow-hidden">
        <div className="container mx-auto max-w-7xl px-4 py-6 h-full">
          <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full h-full flex flex-col">
            {/* Simplified Tab Navigation */}
            <TabsList className="grid w-full grid-cols-3 mb-6 h-10 md:h-12">
              <TabsTrigger value="tasks" className="flex items-center gap-1 md:gap-2 text-xs md:text-sm">
                <CheckCircle className="h-3 w-3 md:h-4 md:w-4" />
                <span>Tasks</span>
              </TabsTrigger>
              <TabsTrigger value="issues" className="flex items-center gap-1 md:gap-2 text-xs md:text-sm">
                <AlertCircle className="h-3 w-3 md:h-4 md:w-4" />
                <span>Issues</span>
              </TabsTrigger>
              <TabsTrigger value="prs" className="flex items-center gap-1 md:gap-2 text-xs md:text-sm">
                <Activity className="h-3 w-3 md:h-4 md:w-4" />
                <span>PRs</span>
              </TabsTrigger>
            </TabsList>

            {/* Tab Content */}
            <div className="flex-1 overflow-y-auto min-h-0">
              <ErrorBoundary
                level="component"
                showDetails={process.env.NODE_ENV === 'development'}
                onError={(error, errorInfo) => {
                  console.error('TaskTab Error:', error, errorInfo);
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
                    </CardContent>
                  </Card>
                )}
              >
                <TaskTab repository={repository} activeTab={activeTab} />
              </ErrorBoundary>
            </div>
          </Tabs>
        </div>
      </div>
    </div>
  );
}