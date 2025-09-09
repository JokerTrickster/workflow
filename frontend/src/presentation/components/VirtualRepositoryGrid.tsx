'use client';

import { memo, useMemo, useCallback } from 'react';
import { FixedSizeList as List } from 'react-window';
import { Repository } from '../../domain/entities/Repository';
import { RepositoryCard } from './RepositoryCard';

interface VirtualRepositoryGridProps {
  repositories: Repository[];
  onSelect: (repository: Repository) => void;
  onConnect: (repoId: number) => void;
  isLoading?: boolean;
  height?: number;
}

interface RowData {
  repositories: Repository[];
  onSelect: (repository: Repository) => void;
  onConnect: (repoId: number) => void;
  isLoading?: boolean;
  itemsPerRow: number;
}

const Row = memo(({ index, style, data }: { index: number; style: any; data: RowData }) => {
  const { repositories, onSelect, onConnect, isLoading, itemsPerRow } = data;
  const startIndex = index * itemsPerRow;
  const endIndex = Math.min(startIndex + itemsPerRow, repositories.length);
  const items = repositories.slice(startIndex, endIndex);

  return (
    <div style={style} className="flex gap-6 px-4">
      {items.map((repository) => (
        <div key={repository.id} className="flex-1 min-w-0">
          <RepositoryCard
            repository={repository}
            onSelect={onSelect}
            onConnect={onConnect}
            isLoading={isLoading}
          />
        </div>
      ))}
      {/* Fill remaining space if needed */}
      {items.length < itemsPerRow && (
        <div className="flex-1 min-w-0" style={{ opacity: 0 }} />
      )}
    </div>
  );
});

Row.displayName = 'VirtualRepositoryRow';

const VirtualRepositoryGrid = memo(function VirtualRepositoryGrid({
  repositories,
  onSelect,
  onConnect,
  isLoading = false,
  height = 600
}: VirtualRepositoryGridProps) {
  // Calculate items per row based on screen size
  const itemsPerRow = useMemo(() => {
    if (typeof window !== 'undefined') {
      const width = window.innerWidth;
      if (width >= 1280) return 2; // xl screens: 2 columns
      if (width >= 1024) return 1; // lg screens: 1 column
      if (width >= 768) return 2;  // md screens: 2 columns
      return 1; // sm screens: 1 column
    }
    return 2; // default for SSR
  }, []);

  // Calculate total number of rows needed
  const rowCount = useMemo(() => {
    return Math.ceil(repositories.length / itemsPerRow);
  }, [repositories.length, itemsPerRow]);

  // Height of each row (card height + gap)
  const ITEM_HEIGHT = 280;

  // Memoize row data to prevent re-renders
  const rowData = useMemo((): RowData => ({
    repositories,
    onSelect,
    onConnect,
    isLoading,
    itemsPerRow
  }), [repositories, onSelect, onConnect, isLoading, itemsPerRow]);

  // Handle window resize to recalculate items per row
  const handleResize = useCallback(() => {
    // Force re-render by changing key when window is resized
    // This is handled by the parent component
  }, []);

  if (repositories.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12" style={{ height }}>
        <p className="text-muted-foreground">No repositories found</p>
      </div>
    );
  }

  return (
    <div className="w-full">
      <List
        height={height}
        itemCount={rowCount}
        itemSize={ITEM_HEIGHT}
        itemData={rowData}
        width="100%"
        className="scrollbar-thin scrollbar-track-gray-100 scrollbar-thumb-gray-300"
      >
        {Row}
      </List>
    </div>
  );
});

export { VirtualRepositoryGrid };