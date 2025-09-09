'use client';

import { memo, useMemo, useCallback } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../components/ui/card';
import { Button } from '../../components/ui/button';
import { Badge } from '../../components/ui/badge';
import { Star, GitFork, Calendar, ExternalLink } from 'lucide-react';

interface RepositoryCardProps {
  repository: Repository;
  onSelect: (repository: Repository) => void;
  onConnect: (repoId: number) => void;
  isLoading?: boolean;
}

const RepositoryCard = memo(function RepositoryCard({ repository, onSelect, onConnect, isLoading }: RepositoryCardProps) {
  // Memoize formatted date to prevent re-computation on every render
  const formattedDate = useMemo(() => 
    new Date(repository.updated_at).toLocaleDateString(), 
    [repository.updated_at]
  );

  // Memoize connect handler to prevent re-renders
  const handleConnect = useCallback(async () => {
    await onConnect(repository.id);
    // After connecting, open the workspace with updated repository
    const updatedRepository = { ...repository, is_connected: true };
    onSelect(updatedRepository);
  }, [onConnect, onSelect, repository]);

  // Memoize select handler
  const handleSelect = useCallback(() => {
    onSelect(repository);
  }, [onSelect, repository]);

  return (
    <Card className="w-full touch-target">
      <CardHeader className="pb-3">
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2">
          <div className="space-y-1 flex-1 min-w-0">
            <CardTitle className="text-base sm:text-lg truncate">{repository.name}</CardTitle>
            <CardDescription className="flex items-center gap-1">
              <a 
                href={repository.html_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-primary truncate touch-target min-h-[28px] flex items-center"
                aria-label={`Open ${repository.full_name} on GitHub`}
              >
                {repository.full_name}
              </a>
              <ExternalLink className="h-3 w-3 shrink-0" />
            </CardDescription>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            {repository.private && (
              <Badge variant="secondary" className="text-xs">Private</Badge>
            )}
            {repository.is_connected && (
              <Badge variant="default" className="text-xs">Connected</Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4 pt-0">
        {repository.description && (
          <p className="text-sm text-muted-foreground line-clamp-2">
            {repository.description}
          </p>
        )}
        
        <div className="flex flex-wrap items-center gap-3 sm:gap-4 text-xs sm:text-sm text-muted-foreground">
          {repository.language && (
            <div className="flex items-center gap-1 shrink-0">
              <div className="w-2 h-2 rounded-full bg-blue-500" />
              <span className="truncate max-w-[60px]">{repository.language}</span>
            </div>
          )}
          <div className="flex items-center gap-1 shrink-0">
            <Star className="h-3 w-3 sm:h-4 sm:w-4" />
            {repository.stargazers_count}
          </div>
          <div className="flex items-center gap-1 shrink-0">
            <GitFork className="h-3 w-3 sm:h-4 sm:w-4" />
            {repository.forks_count}
          </div>
          <div className="flex items-center gap-1 shrink-0">
            <Calendar className="h-3 w-3 sm:h-4 sm:w-4" />
            <span className="hidden sm:inline">{formattedDate}</span>
            <span className="sm:hidden">{new Date(repository.updated_at).toLocaleDateString('ko-KR', { month: 'short', day: 'numeric' })}</span>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          {repository.is_connected ? (
            <Button 
              variant="outline" 
              size="sm"
              onClick={handleSelect}
              disabled={isLoading}
              className="touch-target text-sm"
              aria-label={`Open workspace for ${repository.name}`}
            >
              Open Workspace
            </Button>
          ) : (
            <Button 
              size="sm" 
              onClick={handleConnect}
              disabled={isLoading}
              className="touch-target text-sm"
              aria-label={`Connect ${repository.name} repository`}
            >
              {isLoading ? 'Connecting...' : 'Connect Repository'}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
});

export { RepositoryCard };