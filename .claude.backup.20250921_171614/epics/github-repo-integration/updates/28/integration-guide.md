# Integration Guide: useRepositories Hook

## For Stream A (UI Components)

### Basic Usage

```tsx
import { useRepositories } from '@/hooks/useRepositories';

function RepositoryList() {
  const {
    repositories,
    isLoading,
    isError,
    error,
    connectRepository,
    disconnectRepository,
    isConnecting,
    isDisconnecting
  } = useRepositories();

  if (isLoading) return <div>Loading repositories...</div>;
  if (isError) return <div>Error: {error?.message}</div>;

  return (
    <div>
      {repositories.map(repo => (
        <RepositoryCard
          key={repo.id}
          repository={repo}
          onConnect={(repoId) => connectRepository(repoId)}
          onDisconnect={(repoId) => disconnectRepository(repoId)}
          onOpenWorkspace={(repo) => {
            // Handle workspace opening
          }}
        />
      ))}
    </div>
  );
}
```

### Available Methods & Properties

#### Data & State
- `repositories: Repository[]` - Array of repositories with connection status
- `isLoading: boolean` - Loading state for initial fetch
- `isError: boolean` - Error state
- `error: Error | null` - Error object if any

#### Actions
- `connectRepository(repoId: number, localPath?: string): Promise<void>`
- `disconnectRepository(repoId: number): Promise<void>`
- `refetch(): void` - Manually refetch repositories

#### Action States
- `isConnecting: boolean` - True while any connection is in progress
- `isDisconnecting: boolean` - True while any disconnection is in progress

#### Utilities
- `getConnectionStatus(repoId: number)` - Get specific repo connection info

### Repository Object Structure

```tsx
interface Repository {
  id: number;
  name: string;
  full_name: string;
  description: string | null;
  html_url: string;
  clone_url: string;
  ssh_url: string;
  private: boolean;
  language: string | null;
  stargazers_count: number;
  forks_count: number;
  created_at: string;
  updated_at: string;
  pushed_at: string;
  default_branch: string;
  is_connected?: boolean;  // Added by hook
  local_path?: string;     // Added by hook
}
```

### Connection Status Persistence

Connection states are automatically persisted to localStorage and restored on page refresh. No additional setup required.

### Error Handling

The hook integrates with the existing error handling system. All GitHub API errors and mutations errors are properly handled and exposed through the `isError` and `error` properties.

## Implementation Complete ✅

The `useRepositories` hook is ready for integration and provides all the functionality required for repository connection state management as specified in Issue #28.