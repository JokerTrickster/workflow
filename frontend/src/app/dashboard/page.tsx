'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import { Header } from '../../components/Header';
import { RepositoryCard } from '../../presentation/components/RepositoryCard';
import { WorkspacePanel } from '../../presentation/components/WorkspacePanel';
import { SearchFilter } from '../../presentation/components/SearchFilter';
import { ErrorBoundary } from '../../components/ErrorBoundary';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card';
import { Button } from '../../components/ui/button';
import { Badge } from '../../components/ui/badge';
import { 
  AlertCircle, 
  RefreshCw, 
  Loader2, 
  Github,
  ExternalLink,
  Wifi,
  WifiOff
} from 'lucide-react';
import { useRepositories } from '../../hooks/useRepositories';
import { Repository } from '../../domain/entities/Repository';
import { useNetworkStatus } from '../../hooks/useNetworkStatus';

export default function Dashboard() {
  const [selectedRepository, setSelectedRepository] = useState<Repository | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [languageFilter, setLanguageFilter] = useState<string>('');
  const [connectionFilter, setConnectionFilter] = useState<'all' | 'connected' | 'disconnected'>('all');
  
  // Network status monitoring
  const isOnline = useNetworkStatus();

  // Repository data with error handling
  const {
    repositories,
    isLoading,
    error,
    refetch: refetchRepositories,
    isRefetching,
    connectRepository
  } = useRepositories();
  
  // Debug logging
  console.log('📊 Dashboard state:', {
    repositories: repositories?.length || 0,
    isLoading,
    error: error?.message,
    isRefetching,
    firstRepo: repositories?.[0]?.name
  });

  // Handle repository selection - memoized to prevent re-renders
  const handleRepositorySelect = useCallback((repository: Repository) => {
    setSelectedRepository(repository);
  }, []);

  // Handle closing workspace - memoized
  const handleCloseWorkspace = useCallback(() => {
    setSelectedRepository(null);
  }, []);

  // Handle repository connection - memoized
  const handleRepositoryConnect = useCallback(async (repositoryId: number) => {
    console.log('Connecting repository:', repositoryId);
    try {
      await connectRepository(repositoryId);
      console.log('Repository connected successfully');
    } catch (error) {
      console.error('Failed to connect repository:', error);
    }
  }, [connectRepository]);

  // Filter repositories based on search and filters - memoized for performance
  const filteredRepositories = useMemo(() => {
    if (!repositories) return [];
    
    return repositories.filter((repo) => {
      const matchesSearch = !searchQuery || 
        repo.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        repo.description?.toLowerCase().includes(searchQuery.toLowerCase());
      
      const matchesLanguage = !languageFilter || repo.language === languageFilter;
      
      const matchesConnection = connectionFilter === 'all' || 
        (connectionFilter === 'connected' && repo.is_connected) ||
        (connectionFilter === 'disconnected' && !repo.is_connected);
      
      return matchesSearch && matchesLanguage && matchesConnection;
    });
  }, [repositories, searchQuery, languageFilter, connectionFilter]);
  
  console.log('🔍 Filter results:', {
    totalRepos: repositories?.length || 0,
    filteredCount: filteredRepositories.length,
    searchQuery,
    languageFilter,
    connectionFilter,
    firstFilteredRepo: filteredRepositories[0]?.name
  });

  // Get unique languages for filter - memoized for performance
  const availableLanguages = useMemo(() => {
    if (!repositories) return [];
    
    return Array.from(
      new Set(repositories.map(repo => repo.language).filter(Boolean))
    ).sort();
  }, [repositories]);

  // Handle retry for different error scenarios - memoized
  const handleRetry = useCallback(() => {
    if (!isOnline) {
      // For offline scenarios, just show a message about connectivity
      alert('Please check your internet connection and try again.');
      return;
    }
    
    refetchRepositories();
  }, [isOnline, refetchRepositories]);

  // Show workspace if a repository is selected
  if (selectedRepository) {
    return (
      <ErrorBoundary 
        level="page" 
        showDetails={process.env.NODE_ENV === 'development'}
        fallback={(error, resetError) => (
          <div className="min-h-screen flex items-center justify-center p-4">
            <Card className="w-full max-w-md border-red-200 bg-red-50">
              <CardHeader>
                <CardTitle className="text-red-900 flex items-center gap-2">
                  <AlertCircle className="h-5 w-5" />
                  Workspace Error
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-red-700 mb-4">
                  The workspace panel encountered an error and couldn't load properly.
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
                  <Button 
                    variant="outline" 
                    onClick={() => setSelectedRepository(null)}
                  >
                    Back to Dashboard
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      >
        <WorkspacePanel 
          repository={selectedRepository} 
          onClose={handleCloseWorkspace} 
        />
      </ErrorBoundary>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />
      
      <main className="container mx-auto max-w-7xl px-4 py-8 px-safe-4">
        {/* Network Status Banner */}
        {!isOnline && (
          <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div className="flex items-center gap-2 text-yellow-800">
              <WifiOff className="h-5 w-5" />
              <div>
                <p className="font-medium">You're currently offline</p>
                <p className="text-sm">
                  Some features may be limited. Repository data will be refreshed when connection is restored.
                </p>
              </div>
            </div>
          </div>
        )}

        <div className="flex flex-col lg:flex-row gap-6">
          {/* Sidebar with filters */}
          <div className="lg:w-80 shrink-0">
            <ErrorBoundary 
              level="component"
              fallback={(error, resetError) => (
                <Card className="border-red-200 bg-red-50">
                  <CardHeader>
                    <CardTitle className="text-red-900 flex items-center gap-2">
                      <AlertCircle className="h-5 w-5" />
                      Filter Error
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-red-700 mb-4">
                      The search filters encountered an error.
                    </p>
                    <Button variant="outline" onClick={resetError}>
                      Reset Filters
                    </Button>
                  </CardContent>
                </Card>
              )}
            >
              <SearchFilter
                searchQuery={searchQuery}
                onSearchChange={setSearchQuery}
                languageFilter={languageFilter}
                onLanguageChange={setLanguageFilter}
                connectionFilter={connectionFilter}
                onConnectionFilterChange={setConnectionFilter}
                availableLanguages={availableLanguages}
                isLoading={isLoading}
              />
            </ErrorBoundary>
          </div>

          {/* Main content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h1 className="text-2xl font-bold">Your Repositories</h1>
                <p className="text-sm text-muted-foreground">
                  {isLoading 
                    ? 'Loading repositories...' 
                    : repositories 
                      ? `${filteredRepositories.length} of ${repositories.length} repositories`
                      : 'No repositories found'
                  }
                </p>
              </div>
              
              <div className="flex items-center gap-2">
                {/* Network status indicator */}
                <div className="flex items-center gap-1 text-xs">
                  {isOnline ? (
                    <>
                      <Wifi className="h-3 w-3 text-green-500" />
                      <span className="text-green-600">Online</span>
                    </>
                  ) : (
                    <>
                      <WifiOff className="h-3 w-3 text-red-500" />
                      <span className="text-red-600">Offline</span>
                    </>
                  )}
                </div>
                
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRetry}
                  disabled={isLoading || isRefetching}
                >
                  <RefreshCw className={`h-4 w-4 mr-2 ${(isLoading || isRefetching) ? 'animate-spin' : ''}`} />
                  Refresh
                </Button>
              </div>
            </div>

            {/* Error States */}
            {error && (
              <Card className="border-red-200 bg-red-50 mb-6">
                <CardHeader>
                  <CardTitle className="text-red-900 flex items-center gap-2">
                    <AlertCircle className="h-5 w-5" />
                    Failed to Load Repositories
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <p className="text-red-700 mb-2">
                        {error.message.includes('rate limit') 
                          ? 'GitHub API rate limit exceeded. Please try again later.'
                          : error.message.includes('token') 
                            ? 'GitHub authentication failed. Please sign in again.'
                            : error.message.includes('Network') || error.message.includes('fetch')
                              ? 'Network error occurred. Please check your internet connection.'
                              : 'An unexpected error occurred while loading repositories.'
                        }
                      </p>
                      
                      {error.message.includes('rate limit') && (
                        <div className="text-sm text-red-600 bg-red-100 p-3 rounded">
                          <p className="font-medium">Rate Limit Information:</p>
                          <p>GitHub API limits the number of requests per hour.</p>
                          <p>The limit typically resets every hour. Try again later.</p>
                        </div>
                      )}
                    </div>
                    
                    <div className="flex flex-wrap gap-2">
                      <Button variant="outline" onClick={handleRetry}>
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Try Again
                      </Button>
                      
                      {error.message.includes('token') && (
                        <Button variant="outline" asChild>
                          <a href="/login">
                            <Github className="h-4 w-4 mr-2" />
                            Re-authenticate
                          </a>
                        </Button>
                      )}
                      
                      <Button variant="ghost" asChild>
                        <a 
                          href="https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting" 
                          target="_blank" 
                          rel="noopener noreferrer"
                        >
                          <ExternalLink className="h-4 w-4 mr-2" />
                          GitHub API Docs
                        </a>
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Loading State */}
            {isLoading && !error && (
              <div className="flex flex-col items-center justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin mb-4" />
                <h3 className="text-lg font-semibold mb-2">Loading Repositories</h3>
                <p className="text-sm text-muted-foreground text-center max-w-md">
                  Fetching your GitHub repositories. This may take a few moments if you have many repositories.
                </p>
              </div>
            )}

            {/* Empty State */}
            {!isLoading && !error && repositories && repositories.length === 0 && (
              <div className="flex flex-col items-center justify-center py-12">
                <Github className="h-16 w-16 mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">No Repositories Found</h3>
                <p className="text-sm text-muted-foreground text-center max-w-md mb-4">
                  We couldn't find any repositories in your GitHub account. 
                  Make sure you have repositories and the correct permissions.
                </p>
                <div className="flex gap-2">
                  <Button variant="outline" onClick={handleRetry}>
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Refresh
                  </Button>
                  <Button variant="outline" asChild>
                    <a href="https://github.com/new" target="_blank" rel="noopener noreferrer">
                      <Github className="h-4 w-4 mr-2" />
                      Create Repository
                    </a>
                  </Button>
                </div>
              </div>
            )}

            {/* Filtered Empty State */}
            {!isLoading && !error && repositories && repositories.length > 0 && filteredRepositories.length === 0 && (
              <div className="flex flex-col items-center justify-center py-12">
                <AlertCircle className="h-16 w-16 mb-4 opacity-50" />
                <h3 className="text-lg font-semibold mb-2">No Matching Repositories</h3>
                <p className="text-sm text-muted-foreground text-center max-w-md mb-4">
                  No repositories match your current filters. Try adjusting your search criteria.
                </p>
                <Button 
                  variant="outline" 
                  onClick={() => {
                    setSearchQuery('');
                    setLanguageFilter('');
                    setConnectionFilter('all');
                  }}
                >
                  Clear Filters
                </Button>
              </div>
            )}

            {/* Repository Grid */}
            {!isLoading && !error && filteredRepositories.length > 0 && (
              <ErrorBoundary 
                level="component"
                fallback={(error, resetError) => (
                  <Card className="border-red-200 bg-red-50">
                    <CardHeader>
                      <CardTitle className="text-red-900 flex items-center gap-2">
                        <AlertCircle className="h-5 w-5" />
                        Repository Display Error
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-red-700 mb-4">
                        There was an error displaying the repositories.
                      </p>
                      <div className="flex gap-2">
                        <Button variant="outline" onClick={resetError}>
                          Try Again
                        </Button>
                        <Button variant="outline" onClick={handleRetry}>
                          Reload Data
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                )}
              >
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2 gap-6">
                  {filteredRepositories.map((repository) => (
                    <RepositoryCard
                      key={repository.id}
                      repository={repository}
                      onSelect={handleRepositorySelect}
                      onConnect={handleRepositoryConnect}
                      isLoading={isRefetching}
                    />
                  ))}
                </div>
              </ErrorBoundary>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}