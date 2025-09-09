# Virtual Scrolling Implementation Plan

## Overview
Based on performance analysis showing 58.8-59.4fps (below 60fps target), implementing virtual scrolling is the critical next step to achieve optimal dashboard performance.

## Recommended Library: react-window

### Installation
```bash
npm install react-window react-window-infinite-loader
npm install --save-dev @types/react-window
```

## Implementation Strategy

### 1. Virtual Repository List Component
```typescript
// src/components/VirtualRepositoryList.tsx
import { FixedSizeList as List } from 'react-window';
import { RepositoryCard } from '../presentation/components/RepositoryCard';
import { Repository } from '../domain/entities/Repository';

interface VirtualRepositoryListProps {
  repositories: Repository[];
  onSelect: (repository: Repository) => void;
  onConnect: (repoId: number) => void;
  isLoading?: boolean;
  height?: number;
  itemHeight?: number;
}

export const VirtualRepositoryList = ({
  repositories,
  onSelect,
  onConnect,
  isLoading = false,
  height = 600,
  itemHeight = 180
}: VirtualRepositoryListProps) => {
  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style} className="px-3">
      <RepositoryCard
        repository={repositories[index]}
        onSelect={onSelect}
        onConnect={onConnect}
        isLoading={isLoading}
      />
    </div>
  );

  return (
    <List
      height={height}
      itemCount={repositories.length}
      itemSize={itemHeight}
      itemData={repositories}
      className="repository-virtual-list"
    >
      {Row}
    </List>
  );
};
```

### 2. Dashboard Integration
```typescript
// Update src/app/dashboard/page.tsx
import { VirtualRepositoryList } from '../../components/VirtualRepositoryList';

// Replace the repository grid section with:
{!isLoading && !error && filteredRepositories.length > 0 && (
  <ErrorBoundary 
    level="component"
    fallback={(error, resetError) => (/* Error fallback */)}
  >
    <div className="virtual-list-container">
      <VirtualRepositoryList
        repositories={filteredRepositories}
        onSelect={handleRepositorySelect}
        onConnect={handleRepositoryConnect}
        isLoading={isRefetching}
        height={Math.min(800, window.innerHeight * 0.7)}
        itemHeight={180}
      />
    </div>
  </ErrorBoundary>
)}
```

### 3. Infinite Scrolling Integration
```typescript
// Enhanced virtual list with infinite loading
import InfiniteLoader from 'react-window-infinite-loader';

const VirtualInfiniteRepositoryList = ({
  repositories,
  hasNextPage,
  loadMore,
  ...props
}: VirtualRepositoryListProps & {
  hasNextPage: boolean;
  loadMore: () => Promise<void>;
}) => {
  const itemCount = hasNextPage ? repositories.length + 1 : repositories.length;
  const isItemLoaded = (index: number) => !!repositories[index];

  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => {
    if (!repositories[index]) {
      return (
        <div style={style} className="px-3 flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      );
    }

    return (
      <div style={style} className="px-3">
        <RepositoryCard repository={repositories[index]} {...props} />
      </div>
    );
  };

  return (
    <InfiniteLoader
      isItemLoaded={isItemLoaded}
      itemCount={itemCount}
      loadMoreItems={loadMore}
      threshold={5} // Load more when 5 items away from end
    >
      {({ onItemsRendered, ref }) => (
        <List
          ref={ref}
          height={props.height || 600}
          itemCount={itemCount}
          itemSize={props.itemHeight || 180}
          onItemsRendered={onItemsRendered}
        >
          {Row}
        </List>
      )}
    </InfiniteLoader>
  );
};
```

### 4. IntersectionObserver Hook Integration
```typescript
// Use existing hook for loading trigger
import { useInfiniteScroll } from '../hooks/useIntersectionObserver';

const Dashboard = () => {
  const [page, setPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(true);

  const loadMore = useCallback(async () => {
    if (!hasNextPage) return;
    
    try {
      setPage(prev => prev + 1);
      // Fetch next page of repositories
      const nextRepositories = await fetchRepositories(page + 1);
      if (nextRepositories.length === 0) {
        setHasNextPage(false);
      }
    } catch (error) {
      console.error('Failed to load more repositories:', error);
    }
  }, [page, hasNextPage]);

  return (
    <VirtualInfiniteRepositoryList
      repositories={filteredRepositories}
      hasNextPage={hasNextPage}
      loadMore={loadMore}
      // ... other props
    />
  );
};
```

## Performance Optimizations

### 1. Memoized Row Component
```typescript
const MemoizedRow = memo(({ index, style, data }: {
  index: number;
  style: React.CSSProperties;
  data: {
    repositories: Repository[];
    onSelect: (repo: Repository) => void;
    onConnect: (id: number) => void;
  };
}) => (
  <div style={style} className="px-3">
    <RepositoryCard
      repository={data.repositories[index]}
      onSelect={data.onSelect}
      onConnect={data.onConnect}
    />
  </div>
));
```

### 2. Dynamic Height Support
```typescript
import { VariableSizeList as List } from 'react-window';

const getItemSize = (index: number) => {
  // Calculate based on content
  return repositories[index]?.description?.length > 100 ? 200 : 160;
};
```

### 3. Scrollbar Styling
```css
/* Add to global styles */
.repository-virtual-list {
  scrollbar-width: thin;
  scrollbar-color: #cbd5e0 #f7fafc;
}

.repository-virtual-list::-webkit-scrollbar {
  width: 8px;
}

.repository-virtual-list::-webkit-scrollbar-track {
  background: #f7fafc;
  border-radius: 4px;
}

.repository-virtual-list::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 4px;
}

.repository-virtual-list::-webkit-scrollbar-thumb:hover {
  background: #a0aec0;
}
```

## Testing Strategy

### 1. Performance Validation
```typescript
// Test with performance analyzer
const testVirtualScrolling = async () => {
  const analyzer = new PerformanceAnalyzer();
  const results = await analyzer.runFullAnalysis(500); // Test with 500 repos
  
  console.log('Virtual scrolling performance:', {
    fps: results.scrollMetrics.averageFps,
    memory: results.memoryMetrics.usedJSHeapSize,
    target: results.scrollMetrics.averageFps >= 60 ? 'PASSED' : 'FAILED'
  });
};
```

### 2. E2E Test Updates
```typescript
// Add to existing tests
test('virtual scrolling maintains 60fps with 200+ repositories', async ({ page }) => {
  await page.goto('/dashboard');
  
  // Inject test data and measure performance
  const fps = await page.evaluate(testVirtualScrolling);
  expect(fps.averageFps).toBeGreaterThanOrEqual(60);
});
```

## Migration Plan

### Phase 1: Implementation (Week 1)
1. Install react-window dependencies
2. Create VirtualRepositoryList component
3. Update Dashboard to use virtual scrolling
4. Basic performance validation

### Phase 2: Enhancement (Week 2)
1. Add infinite scrolling with InfiniteLoader
2. Integrate with existing IntersectionObserver hook
3. Implement loading states and error handling
4. Performance optimization and testing

### Phase 3: Polish (Week 3)
1. Add dynamic height support
2. Implement smooth scrolling behaviors
3. Add accessibility improvements
4. Final performance validation and monitoring setup

## Expected Performance Gains

Based on virtual scrolling best practices:
- **FPS Improvement**: +5-15fps (target: >60fps)
- **Memory Reduction**: 60-80% for large lists
- **Initial Render Time**: 70-90% faster
- **Scroll Responsiveness**: Near-instant scroll initiation

## Rollback Plan

If virtual scrolling introduces issues:
1. Feature flag implementation allows instant rollback
2. Keep original grid implementation as fallback
3. A/B testing to compare performance metrics

---

This implementation plan provides a clear path to achieve the 60fps target while maintaining the existing functionality and user experience.