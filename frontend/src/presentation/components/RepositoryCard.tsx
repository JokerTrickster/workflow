'use client';

import { memo, useMemo, useCallback } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../components/ui/card';
import { Button } from '../../components/ui/button';
import { Badge } from '../../components/ui/badge';
import { Star, GitFork, ExternalLink } from 'lucide-react';

// Utility functions for enhanced UI
const getLanguageColor = (language: string): string => {
  const colors: Record<string, string> = {
    'TypeScript': 'bg-blue-500',
    'JavaScript': 'bg-yellow-500',
    'Python': 'bg-green-500',
    'Java': 'bg-orange-500',
    'Go': 'bg-cyan-500',
    'Rust': 'bg-orange-600',
    'PHP': 'bg-purple-500',
    'Ruby': 'bg-red-500',
    'Swift': 'bg-orange-500',
    'Kotlin': 'bg-purple-600',
    'C++': 'bg-blue-600',
    'C#': 'bg-purple-700',
    'HTML': 'bg-orange-400',
    'CSS': 'bg-blue-400',
    'Vue': 'bg-green-400',
    'React': 'bg-cyan-400',
    'Angular': 'bg-red-600',
  };
  return colors[language] || 'bg-gray-500';
};

const formatCount = (count: number): string => {
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
  if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
  return count.toString();
};

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
    try {
      await onConnect(repository.id);
      // Don't auto-select after connecting - let user choose
    } catch (error) {
      console.error('Failed to connect repository:', error);
    }
  }, [onConnect, repository.id]);

  // Memoize select handler
  const handleSelect = useCallback(() => {
    onSelect(repository);
  }, [onSelect, repository]);

  return (
    <Card className="h-full flex flex-col transition-all duration-200 hover:shadow-md">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="space-y-1 flex-1 min-w-0">
            <CardTitle className="text-lg font-semibold truncate">{repository.name}</CardTitle>
            <CardDescription className="flex items-center gap-1 text-sm">
              <a 
                href={repository.html_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-primary transition-colors truncate flex items-center gap-1"
              >
                {repository.full_name}
                <ExternalLink className="h-3 w-3" />
              </a>
            </CardDescription>
          </div>
          
          <div className="flex items-center gap-2 ml-2">
            {repository.private && (
              <Badge variant="secondary" className="text-xs">Private</Badge>
            )}
            {repository.is_connected && (
              <Badge variant="default" className="text-xs bg-green-100 text-green-700 border-green-200">
                Connected
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="flex-1 flex flex-col justify-between space-y-4 pt-0">
        <div className="space-y-3">
          {repository.description && (
            <p className="text-sm text-muted-foreground leading-relaxed overflow-hidden" style={{
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              maxHeight: '2.5rem'
            }}>
              {repository.description}
            </p>
          )}
          
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <div className="flex items-center gap-3">
              {repository.language && (
                <div className="flex items-center gap-1">
                  <div className={`w-2 h-2 rounded-full ${getLanguageColor(repository.language)}`} />
                  {repository.language}
                </div>
              )}
              <div className="flex items-center gap-1">
                <Star className="h-4 w-4 text-yellow-500" />
                {formatCount(repository.stargazers_count)}
              </div>
              <div className="flex items-center gap-1">
                <GitFork className="h-4 w-4" />
                {formatCount(repository.forks_count)}
              </div>
            </div>
            <div className="text-xs">
              {formattedDate}
            </div>
          </div>
        </div>

        <div className="pt-3 mt-auto">
          {repository.is_connected ? (
            <Button 
              className="w-full"
              variant="default"
              size="sm"
              onClick={handleSelect}
              disabled={isLoading}
            >
              Open Workspace
            </Button>
          ) : (
            <Button 
              className="w-full"
              variant="outline"
              size="sm"
              onClick={handleConnect}
              disabled={isLoading}
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