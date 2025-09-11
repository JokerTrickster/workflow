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
    <Card className="group relative w-full h-full transition-all duration-300 hover:shadow-lg hover:shadow-primary/10 hover:-translate-y-1 border-0 bg-gradient-to-br from-card via-card to-muted/30">
      {/* Background Pattern */}
      <div className="absolute inset-0 bg-grid-small opacity-[0.02] rounded-lg" />
      
      {/* Connection Status Indicator */}
      <div className={`absolute top-0 right-0 w-3 h-3 rounded-bl-lg transition-colors duration-300 ${
        repository.is_connected ? 'bg-green-500' : repository.private ? 'bg-orange-500' : 'bg-blue-500'
      }`} />
      
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="space-y-2 flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <CardTitle className="text-lg font-bold bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text truncate">
                {repository.name}
              </CardTitle>
              {repository.private && (
                <Badge variant="secondary" className="text-xs px-2 py-0.5 bg-orange-100 text-orange-700 border-orange-200">
                  Private
                </Badge>
              )}
            </div>
            <CardDescription className="flex items-center gap-1.5 text-xs">
              <a 
                href={repository.html_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-primary transition-colors truncate flex items-center gap-1"
              >
                {repository.full_name}
                <ExternalLink className="h-3 w-3 opacity-60" />
              </a>
            </CardDescription>
          </div>
          
          {repository.is_connected && (
            <Badge variant="default" className="ml-2 bg-green-100 text-green-700 border-green-200 animate-pulse">
              Connected
            </Badge>
          )}
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4 pt-0">
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
        
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            {repository.language && (
              <div className="flex items-center gap-1.5">
                <div className={`w-2.5 h-2.5 rounded-full shadow-sm ${getLanguageColor(repository.language)}`} />
                <span className="font-medium">{repository.language}</span>
              </div>
            )}
            <div className="flex items-center gap-1">
              <Star className="h-3.5 w-3.5 text-yellow-500" />
              <span>{formatCount(repository.stargazers_count)}</span>
            </div>
            <div className="flex items-center gap-1">
              <GitFork className="h-3.5 w-3.5" />
              <span>{formatCount(repository.forks_count)}</span>
            </div>
          </div>
          
          <div className="text-xs text-muted-foreground">
            {formattedDate}
          </div>
        </div>

        <div className="pt-2 border-t border-border/50">
          {repository.is_connected ? (
            <Button 
              className="w-full bg-gradient-to-r from-primary to-primary/90 hover:from-primary/90 hover:to-primary text-white shadow-sm transition-all duration-200 group-hover:shadow-md"
              size="sm"
              onClick={handleSelect}
              disabled={isLoading}
            >
              <span className="flex items-center gap-2">
                Open Workspace
                <ExternalLink className="h-3.5 w-3.5 opacity-70" />
              </span>
            </Button>
          ) : (
            <Button 
              className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white shadow-sm transition-all duration-200 group-hover:shadow-md"
              size="sm"
              onClick={handleConnect}
              disabled={isLoading}
            >
              {isLoading ? (
                <span className="flex items-center gap-2">
                  <div className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Connecting...
                </span>
              ) : (
                'Connect Repository'
              )}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
});

export { RepositoryCard };