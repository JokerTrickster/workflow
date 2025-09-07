'use client'

import { useRepositories } from '@/hooks/useRepositories'
import { useAuth } from '@/contexts/AuthContext'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { RefreshCw, GitBranch, Star, GitFork, Calendar, ExternalLink } from 'lucide-react'
import { format } from 'date-fns'

export function RepositoriesTest() {
  const { isAuthenticated } = useAuth()
  const {
    repositories,
    loading,
    error,
    totalCount,
    hasNextPage,
    rateLimit,
    fetchRepositories,
    fetchAllRepositories,
    refresh,
  } = useRepositories()

  if (!isAuthenticated) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-center text-muted-foreground">
            Please log in to test GitHub API integration
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* API Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <GitBranch className="w-5 h-5" />
            GitHub API Test
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Control buttons */}
          <div className="flex flex-wrap gap-2">
            <Button 
              onClick={() => fetchRepositories()} 
              disabled={loading}
              className="gap-2"
            >
              {loading ? <LoadingSpinner size="sm" /> : <RefreshCw className="w-4 h-4" />}
              Fetch Repositories
            </Button>
            
            <Button 
              onClick={() => fetchAllRepositories(3)} 
              disabled={loading}
              variant="outline"
              className="gap-2"
            >
              Fetch All (3 pages max)
            </Button>
            
            <Button 
              onClick={refresh} 
              disabled={loading}
              variant="outline"
              size="sm"
            >
              Refresh
            </Button>
          </div>

          {/* Status info */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <div className="font-medium">Total Repos</div>
              <div className="text-muted-foreground">{totalCount}</div>
            </div>
            <div>
              <div className="font-medium">Loaded</div>
              <div className="text-muted-foreground">{repositories.length}</div>
            </div>
            <div>
              <div className="font-medium">Has More</div>
              <div className="text-muted-foreground">{hasNextPage ? 'Yes' : 'No'}</div>
            </div>
            <div>
              <div className="font-medium">Rate Limit</div>
              <div className="text-muted-foreground">
                {rateLimit ? `${rateLimit.remaining} remaining` : 'Unknown'}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Error display */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Loading state */}
      {loading && (
        <Card>
          <CardContent className="pt-6">
            <LoadingSpinner message="Fetching repositories from GitHub..." />
          </CardContent>
        </Card>
      )}

      {/* Repositories grid */}
      {repositories.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">
            Repositories ({repositories.length})
          </h3>
          
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {repositories.slice(0, 12).map((repo) => (
              <Card key={repo.id} className="hover:shadow-md transition-shadow">
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="min-w-0 flex-1">
                      <CardTitle className="text-base truncate">
                        {repo.name}
                      </CardTitle>
                      <p className="text-sm text-muted-foreground truncate">
                        {repo.fullName}
                      </p>
                    </div>
                    <div className="flex gap-1 ml-2">
                      {repo.isPrivate && (
                        <Badge variant="secondary" className="text-xs">Private</Badge>
                      )}
                      {repo.isFork && (
                        <Badge variant="outline" className="text-xs">Fork</Badge>
                      )}
                      {repo.isArchived && (
                        <Badge variant="destructive" className="text-xs">Archived</Badge>
                      )}
                    </div>
                  </div>
                </CardHeader>
                
                <CardContent className="space-y-3">
                  {repo.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {repo.description}
                    </p>
                  )}
                  
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    {repo.language && (
                      <div className="flex items-center gap-1">
                        <div className="w-3 h-3 rounded-full bg-blue-500" />
                        <span>{repo.language}</span>
                      </div>
                    )}
                    
                    {repo.starCount > 0 && (
                      <div className="flex items-center gap-1">
                        <Star className="w-3 h-3" />
                        <span>{repo.starCount}</span>
                      </div>
                    )}
                    
                    {repo.forkCount > 0 && (
                      <div className="flex items-center gap-1">
                        <GitFork className="w-3 h-3" />
                        <span>{repo.forkCount}</span>
                      </div>
                    )}
                  </div>
                  
                  {repo.updatedAt && (
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Calendar className="w-3 h-3" />
                      <span>Updated {format(new Date(repo.updatedAt), 'MMM d, yyyy')}</span>
                    </div>
                  )}
                  
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      className="gap-1"
                      asChild
                    >
                      <a 
                        href={repo.url} 
                        target="_blank" 
                        rel="noopener noreferrer"
                      >
                        <ExternalLink className="w-3 h-3" />
                        View
                      </a>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
          
          {repositories.length > 12 && (
            <div className="text-center text-sm text-muted-foreground">
              Showing first 12 of {repositories.length} repositories
            </div>
          )}
        </div>
      )}
      
      {!loading && repositories.length === 0 && !error && (
        <Card>
          <CardContent className="pt-6 text-center">
            <p className="text-muted-foreground">
              No repositories found. Click "Fetch Repositories" to load data.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}