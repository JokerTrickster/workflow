'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { ErrorHandler } from '../../utils/errorHandler';
import { useErrorRecovery } from '../../hooks/useErrorRecovery';
import { useAuth } from '../../contexts/AuthContext';
import { ErrorMessage } from '../../components/ErrorMessage';
import { ErrorTestPanel } from '../../components/ErrorTestPanel';
import { RepositoryCard } from '../../presentation/components/RepositoryCard';
import { SearchFilter } from '../../presentation/components/SearchFilter';
import { WorkspacePanel } from '../../presentation/components/WorkspacePanel';
import { Button } from '../../components/ui/button';
import { Card, CardContent } from '../../components/ui/card';
import { Github, Plus, RefreshCw } from 'lucide-react';
import { Repository } from '../../domain/entities/Repository';
import { GitHubApiService } from '../../services/githubApi';

// Mock data for development - TODO: Replace with actual API call
const mockRepositories: Repository[] = [
  {
    id: 1,
    name: 'ai-git-workbench',
    full_name: 'captain/ai-git-workbench',
    description: 'AI-powered Git workbench for managing multiple repositories',
    html_url: 'https://github.com/captain/ai-git-workbench',
    clone_url: 'https://github.com/captain/ai-git-workbench.git',
    ssh_url: 'git@github.com:captain/ai-git-workbench.git',
    private: false,
    language: 'TypeScript',
    stargazers_count: 42,
    forks_count: 8,
    created_at: '2024-01-15T10:30:00Z',
    updated_at: '2024-09-06T08:15:00Z',
    pushed_at: '2024-09-06T08:15:00Z',
    default_branch: 'main',
    is_connected: false,
  },
  {
    id: 2,
    name: 'workflow',
    full_name: 'captain/workflow',
    description: 'Workflow management system with Next.js and Go backend',
    html_url: 'https://github.com/captain/workflow',
    clone_url: 'https://github.com/captain/workflow.git',
    ssh_url: 'git@github.com:captain/workflow.git',
    private: false,
    language: 'TypeScript',
    stargazers_count: 25,
    forks_count: 5,
    created_at: '2024-02-01T09:00:00Z',
    updated_at: '2024-09-08T10:30:00Z',
    pushed_at: '2024-09-08T10:30:00Z',
    default_branch: 'main',
    is_connected: true,
  },
  {
    id: 3,
    name: 'sample-project',
    full_name: 'captain/sample-project',
    description: 'A sample project for testing various features',
    html_url: 'https://github.com/captain/sample-project',
    clone_url: 'https://github.com/captain/sample-project.git',
    ssh_url: 'git@github.com:captain/sample-project.git',
    private: true,
    language: 'JavaScript',
    stargazers_count: 12,
    forks_count: 3,
    created_at: '2024-02-20T14:20:00Z',
    updated_at: '2024-09-05T16:45:00Z',
    pushed_at: '2024-09-05T16:45:00Z',
    default_branch: 'main',
    is_connected: true,
  },
  {
    id: 4,
    name: 'react-components',
    full_name: 'captain/react-components',
    description: 'Reusable React components library',
    html_url: 'https://github.com/captain/react-components',
    clone_url: 'https://github.com/captain/react-components.git',
    ssh_url: 'git@github.com:captain/react-components.git',
    private: false,
    language: 'TypeScript',
    stargazers_count: 89,
    forks_count: 15,
    created_at: '2024-03-10T11:15:00Z',
    updated_at: '2024-09-07T14:20:00Z',
    pushed_at: '2024-09-07T14:20:00Z',
    default_branch: 'main',
    is_connected: false,
  },
  {
    id: 5,
    name: 'python-scripts',
    full_name: 'captain/python-scripts',
    description: 'Collection of useful Python automation scripts',
    html_url: 'https://github.com/captain/python-scripts',
    clone_url: 'https://github.com/captain/python-scripts.git',
    ssh_url: 'git@github.com:captain/python-scripts.git',
    private: false,
    language: 'Python',
    stargazers_count: 67,
    forks_count: 12,
    created_at: '2024-04-05T16:30:00Z',
    updated_at: '2024-09-04T18:45:00Z',
    pushed_at: '2024-09-04T18:45:00Z',
    default_branch: 'main',
    is_connected: false,
  },
];

interface SearchFilters {
  query: string;
  language: string;
  visibility: 'all' | 'public' | 'private';
  connected: 'all' | 'connected' | 'disconnected';
}

export default function DashboardPage() {
  const [selectedRepository, setSelectedRepository] = useState<Repository | null>(null);
  const [filters, setFilters] = useState<SearchFilters>({
    query: '',
    language: '',
    visibility: 'all',
    connected: 'all',
  });

  const { error: recoveryError, setError, clearError } = useErrorRecovery();
  const { user, isAuthenticated } = useAuth();

  // Infinite query for repositories with pagination using GitHub API
  const {
    data,
    isLoading,
    error,
    refetch,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ['repositories', user?.id],
    queryFn: async ({ pageParam = 1 }) => {
      console.log('Fetching repositories from GitHub API, page:', pageParam);
      
      if (!isAuthenticated) {
        throw new Error('User not authenticated');
      }

      try {
        const result = await GitHubApiService.fetchUserRepositories(pageParam, 30);
        console.log('GitHub API result:', result);
        return result;
      } catch (error) {
        console.error('GitHub API error:', error);
        throw error;
      }
    },
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    initialPageParam: 1,
    enabled: isAuthenticated, // Only run query when user is authenticated
  });

  // Flatten all pages of repositories
  const allRepositories = useMemo(() => {
    return data?.pages.flatMap(page => page.repositories) || [];
  }, [data]);

  const [localRepositories, setLocalRepositories] = useState(allRepositories);

  // Update local state when query data changes
  useEffect(() => {
    setLocalRepositories(allRepositories);
  }, [allRepositories]);

  const filteredRepositories = useMemo(() => {
    return localRepositories.filter(repo => {
      const matchesQuery = 
        filters.query === '' ||
        repo.name.toLowerCase().includes(filters.query.toLowerCase()) ||
        repo.description?.toLowerCase().includes(filters.query.toLowerCase()) ||
        repo.full_name.toLowerCase().includes(filters.query.toLowerCase());

      const matchesLanguage =
        filters.language === '' ||
        repo.language?.toLowerCase() === filters.language.toLowerCase();

      const matchesVisibility =
        filters.visibility === 'all' ||
        (filters.visibility === 'public' && !repo.private) ||
        (filters.visibility === 'private' && repo.private);

      const matchesConnected =
        filters.connected === 'all' ||
        (filters.connected === 'connected' && repo.is_connected) ||
        (filters.connected === 'disconnected' && !repo.is_connected);

      return matchesQuery && matchesLanguage && matchesVisibility && matchesConnected;
    });
  }, [localRepositories, filters]);

  const handleConnectRepository = (repoId: number) => {
    setLocalRepositories(prev => 
      prev.map(repo => 
        repo.id === repoId 
          ? { ...repo, is_connected: true }
          : repo
      )
    );
  };

  const handleOpenWorkspace = (repository: Repository) => {
    setSelectedRepository(repository);
  };

  const handleCloseWorkspace = () => {
    setSelectedRepository(null);
  };

  const handleFiltersChange = (newFilters: SearchFilters) => {
    setFilters(newFilters);
  };

  const handleRefresh = () => {
    clearError(); // 에러 상태 초기화
    refetch();
  };

  // 에러 처리
  useEffect(() => {
    if (error) {
      const appError = ErrorHandler.fromHttpError(error);
      setError(appError, 'Dashboard - Repository fetch');
    }
  }, [error, setError]);

  // Infinite scroll handler
  const handleScroll = useCallback(() => {
    if (
      window.innerHeight + document.documentElement.scrollTop
      !== document.documentElement.offsetHeight
    ) {
      return;
    }
    
    if (hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  useEffect(() => {
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [handleScroll]);

  // Authentication check
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-background">
        <header className="border-b">
          <div className="container mx-auto max-w-7xl px-4 py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Github className="h-8 w-8" />
                <h1 className="text-2xl font-bold">Repository Dashboard</h1>
              </div>
            </div>
          </div>
        </header>
        <div className="container mx-auto max-w-7xl px-4 py-8">
          <Card className="max-w-md mx-auto">
            <CardContent className="pt-6 text-center space-y-4">
              <div>
                <Github className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <h3 className="text-lg font-semibold">Sign in required</h3>
                <p className="text-sm text-muted-foreground">
                  Please sign in with GitHub to view your repositories
                </p>
              </div>
              <Button 
                onClick={() => window.location.href = '/'}
                className="gap-2"
              >
                <Github className="h-4 w-4" />
                Go to Sign In
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // 에러 상태 표시
  if (recoveryError) {
    return (
      <div className="min-h-screen bg-background">
        <header className="border-b">
          <div className="container mx-auto max-w-7xl px-4 py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Github className="h-8 w-8" />
                <h1 className="text-2xl font-bold">Repository Dashboard</h1>
              </div>
            </div>
          </div>
        </header>
        <div className="container mx-auto max-w-7xl px-4 py-8">
          <div className="max-w-md mx-auto">
            <ErrorMessage
              error={recoveryError}
              onRetry={handleRefresh}
              showDetails={process.env.NODE_ENV === 'development'}
            />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto max-w-7xl px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Github className="h-8 w-8" />
              <h1 className="text-2xl font-bold">Repository Dashboard</h1>
            </div>
            <div className="flex items-center gap-4">
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefresh}
                disabled={isLoading}
                className="gap-2"
              >
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                Refresh
              </Button>
              <Button variant="outline" size="sm">
                Settings
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto max-w-7xl px-4 py-8">
        <div className="space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h2 className="text-3xl font-bold">Your Repositories</h2>
              <p className="text-muted-foreground">
                {isLoading 
                  ? 'Loading repositories...' 
                  : `${filteredRepositories.length} of ${localRepositories.length} repositories`
                }
              </p>
            </div>
          </div>

          <SearchFilter
            filters={filters}
            onFiltersChange={handleFiltersChange}
            repositories={localRepositories}
          />

          {isLoading ? (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 6 }).map((_, index) => (
                <Card key={index} className="w-full animate-pulse">
                  <CardContent className="p-6 space-y-4">
                    <div className="h-4 bg-muted rounded w-3/4"></div>
                    <div className="h-3 bg-muted rounded w-1/2"></div>
                    <div className="space-y-2">
                      <div className="h-3 bg-muted rounded"></div>
                      <div className="h-3 bg-muted rounded w-5/6"></div>
                    </div>
                    <div className="flex justify-between items-center">
                      <div className="h-3 bg-muted rounded w-1/3"></div>
                      <div className="h-8 bg-muted rounded w-20"></div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : filteredRepositories.length === 0 ? (
            <Card className="max-w-md mx-auto">
              <CardContent className="pt-6 text-center space-y-4">
                <div>
                  <Github className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <h3 className="text-lg font-semibold">
                    {localRepositories.length === 0 
                      ? 'No repositories found'
                      : 'No repositories match your filters'
                    }
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    {localRepositories.length === 0
                      ? 'Connect to GitHub to see your repositories'
                      : 'Try adjusting your search criteria'
                    }
                  </p>
                </div>
                {filters.query || filters.language || filters.visibility !== 'all' || filters.connected !== 'all' ? (
                  <Button
                    variant="outline"
                    onClick={() => setFilters({
                      query: '',
                      language: '',
                      visibility: 'all',
                      connected: 'all',
                    })}
                  >
                    Clear Filters
                  </Button>
                ) : (
                  <Button className="gap-2">
                    <Plus className="h-4 w-4" />
                    Connect GitHub
                  </Button>
                )}
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-6">
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {filteredRepositories.map((repository) => (
                  <RepositoryCard
                    key={repository.id}
                    repository={repository}
                    onConnect={handleConnectRepository}
                    onOpenWorkspace={handleOpenWorkspace}
                  />
                ))}
              </div>

              {/* Load More / Loading State */}
              {hasNextPage && (
                <div className="flex justify-center pt-6">
                  <Button
                    onClick={() => fetchNextPage()}
                    disabled={isFetchingNextPage}
                    variant="outline"
                    className="gap-2"
                  >
                    <RefreshCw className={`h-4 w-4 ${isFetchingNextPage ? 'animate-spin' : ''}`} />
                    {isFetchingNextPage ? 'Loading more...' : 'Load more repositories'}
                  </Button>
                </div>
              )}

              {/* End of results indicator */}
              {!hasNextPage && localRepositories.length > 0 && (
                <div className="text-center text-muted-foreground text-sm py-6">
                  You've reached the end of your repositories
                </div>
              )}
            </div>
          )}
        </div>
      </main>
      
      {/* Workspace Panel */}
      {selectedRepository && (
        <WorkspacePanel
          repository={selectedRepository}
          onClose={handleCloseWorkspace}
        />
      )}
      
      {/* Error Testing Panel - Development Only */}
      <ErrorTestPanel />
    </div>
  );
}