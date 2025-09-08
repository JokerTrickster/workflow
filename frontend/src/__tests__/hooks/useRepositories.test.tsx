import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useRepositories } from '../../hooks/useRepositories';
import { GitHubApiService } from '../../services/githubApi';
import { Repository } from '../../domain/entities/Repository';
import { ReactNode } from 'react';

// Mock the GitHubApiService
jest.mock('../../services/githubApi', () => ({
  GitHubApiService: {
    fetchUserRepositories: jest.fn(),
  },
}));

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

const mockRepositories: Repository[] = [
  {
    id: 1,
    name: 'test-repo',
    full_name: 'user/test-repo',
    description: 'A test repository',
    html_url: 'https://github.com/user/test-repo',
    clone_url: 'https://github.com/user/test-repo.git',
    ssh_url: 'git@github.com:user/test-repo.git',
    private: false,
    language: 'TypeScript',
    stargazers_count: 5,
    forks_count: 2,
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    pushed_at: '2024-01-01T00:00:00Z',
    default_branch: 'main',
    is_connected: false,
  },
  {
    id: 2,
    name: 'another-repo',
    full_name: 'user/another-repo',
    description: null,
    html_url: 'https://github.com/user/another-repo',
    clone_url: 'https://github.com/user/another-repo.git',
    ssh_url: 'git@github.com:user/another-repo.git',
    private: true,
    language: 'JavaScript',
    stargazers_count: 10,
    forks_count: 1,
    created_at: '2023-06-01T00:00:00Z',
    updated_at: '2024-01-15T00:00:00Z',
    pushed_at: '2024-01-15T00:00:00Z',
    default_branch: 'main',
    is_connected: false,
  },
];

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useRepositories', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.getItem.mockReturnValue(null);
    (GitHubApiService.fetchUserRepositories as jest.Mock).mockResolvedValue({
      repositories: mockRepositories,
      hasMore: false,
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should fetch repositories successfully', async () => {
    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isLoading).toBe(true);
    expect(result.current.repositories).toEqual([]);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.repositories).toHaveLength(2);
    expect(result.current.repositories[0].name).toBe('test-repo');
    expect(result.current.repositories[1].name).toBe('another-repo');
    expect(result.current.isError).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it('should handle GitHub API errors', async () => {
    const mockError = new Error('GitHub API Error');
    (GitHubApiService.fetchUserRepositories as jest.Mock).mockRejectedValue(mockError);

    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toBe(mockError);
    expect(result.current.repositories).toEqual([]);
  });

  it('should connect a repository successfully', async () => {
    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    await act(async () => {
      await result.current.connectRepository(1, '/local/path');
    });

    // Check that localStorage was called with correct data
    expect(localStorageMock.setItem).toHaveBeenCalledWith(
      'repository_connections',
      expect.stringContaining('"1"')
    );

    // Verify connection status
    const connectionStatus = result.current.getConnectionStatus(1);
    expect(connectionStatus.is_connected).toBe(true);
    expect(connectionStatus.local_path).toBe('/local/path');
    expect(connectionStatus.connected_at).toBeDefined();
  });

  it('should disconnect a repository successfully', async () => {
    // Mock existing connection state
    const existingConnections = {
      '1': {
        is_connected: true,
        local_path: '/local/path',
        connected_at: '2024-01-01T00:00:00Z',
      },
    };
    localStorageMock.getItem.mockReturnValue(JSON.stringify(existingConnections));

    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Verify initial connected state
    let connectionStatus = result.current.getConnectionStatus(1);
    expect(connectionStatus.is_connected).toBe(true);

    await act(async () => {
      await result.current.disconnectRepository(1);
    });

    // Check that localStorage was updated
    expect(localStorageMock.setItem).toHaveBeenCalledWith(
      'repository_connections',
      expect.stringContaining('"is_connected":false')
    );

    // Verify disconnection status
    connectionStatus = result.current.getConnectionStatus(1);
    expect(connectionStatus.is_connected).toBe(false);
    expect(connectionStatus.local_path).toBeUndefined();
  });

  it('should load connection states from localStorage on mount', async () => {
    const existingConnections = {
      '1': {
        is_connected: true,
        local_path: '/existing/path',
        connected_at: '2024-01-01T00:00:00Z',
      },
    };
    localStorageMock.getItem.mockReturnValue(JSON.stringify(existingConnections));

    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    const connectionStatus = result.current.getConnectionStatus(1);
    expect(connectionStatus.is_connected).toBe(true);
    expect(connectionStatus.local_path).toBe('/existing/path');
    expect(connectionStatus.connected_at).toBe('2024-01-01T00:00:00Z');

    // Check that repository data is merged with connection state
    const connectedRepo = result.current.repositories.find(repo => repo.id === 1);
    expect(connectedRepo?.is_connected).toBe(true);
    expect(connectedRepo?.local_path).toBe('/existing/path');
  });

  it('should handle localStorage errors gracefully', async () => {
    localStorageMock.getItem.mockImplementation(() => {
      throw new Error('localStorage error');
    });

    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Should still work without localStorage
    expect(result.current.repositories).toHaveLength(2);
    expect(result.current.repositories.every(repo => !repo.is_connected)).toBe(true);
  });

  it('should provide refetch functionality', async () => {
    const { result } = renderHook(() => useRepositories(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(GitHubApiService.fetchUserRepositories).toHaveBeenCalledTimes(1);

    act(() => {
      result.current.refetch();
    });

    expect(GitHubApiService.fetchUserRepositories).toHaveBeenCalledTimes(2);
  });
});