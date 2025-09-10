'use client';

import { memo } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { RepositoryCard } from './RepositoryCard';

// Simple grid without virtualization for better mobile compatibility
const SimpleRepositoryGrid = memo(function SimpleRepositoryGrid({
  repositories,
  onSelect,
  onConnect,
  isLoading = false
}: {
  repositories: Repository[];
  onSelect: (repository: Repository) => void;
  onConnect: (repoId: number) => void;
  isLoading?: boolean;
}) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
      {repositories.map((repository) => (
        <RepositoryCard
          key={repository.id}
          repository={repository}
          onSelect={onSelect}
          onConnect={onConnect}
          isLoading={isLoading}
        />
      ))}
    </div>
  );
});

interface VirtualRepositoryGridProps {
  repositories: Repository[];
  onSelect: (repository: Repository) => void;
  onConnect: (repoId: number) => void;
  isLoading?: boolean;
  height?: number;
}

const VirtualRepositoryGrid = memo(function VirtualRepositoryGrid({
  repositories,
  onSelect,
  onConnect,
  isLoading = false,
  height = 600
}: VirtualRepositoryGridProps) {
  if (repositories.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12" style={{ height }}>
        <p className="text-muted-foreground">No repositories found</p>
      </div>
    );
  }

  // For mobile optimization, use a simple grid instead of virtualization
  // This provides better scrolling performance on mobile devices
  return (
    <div className="w-full">
      <SimpleRepositoryGrid
        repositories={repositories}
        onSelect={onSelect}
        onConnect={onConnect}
        isLoading={isLoading}
      />
    </div>
  );
});

export { VirtualRepositoryGrid };