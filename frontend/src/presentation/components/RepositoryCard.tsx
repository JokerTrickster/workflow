'use client';

import { Repository } from '../../domain/entities/Repository';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../components/ui/card';
import { Button } from '../../components/ui/button';
import { Badge } from '../../components/ui/badge';
import { Star, GitFork, Calendar, ExternalLink } from 'lucide-react';

interface RepositoryCardProps {
  repository: Repository;
  onConnect: (repoId: number) => void;
  onOpenWorkspace: (repository: Repository) => void;
}

export function RepositoryCard({ repository, onConnect, onOpenWorkspace }: RepositoryCardProps) {
  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="text-lg">{repository.name}</CardTitle>
            <CardDescription className="flex items-center gap-1">
              <a 
                href={repository.html_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-primary"
              >
                {repository.full_name}
              </a>
              <ExternalLink className="h-3 w-3" />
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {repository.private && (
              <Badge variant="secondary">Private</Badge>
            )}
            {repository.is_connected && (
              <Badge variant="default">Connected</Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {repository.description && (
          <p className="text-sm text-muted-foreground">
            {repository.description}
          </p>
        )}
        
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          {repository.language && (
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-blue-500" />
              {repository.language}
            </div>
          )}
          <div className="flex items-center gap-1">
            <Star className="h-4 w-4" />
            {repository.stargazers_count}
          </div>
          <div className="flex items-center gap-1">
            <GitFork className="h-4 w-4" />
            {repository.forks_count}
          </div>
          <div className="flex items-center gap-1">
            <Calendar className="h-4 w-4" />
            {new Date(repository.updated_at).toLocaleDateString()}
          </div>
        </div>

        <div className="flex justify-end">
          {repository.is_connected ? (
            <Button 
              variant="outline" 
              size="sm"
              onClick={() => onOpenWorkspace(repository)}
            >
              Open Workspace
            </Button>
          ) : (
            <Button 
              size="sm" 
              onClick={() => onConnect(repository.id)}
            >
              Connect Repository
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}