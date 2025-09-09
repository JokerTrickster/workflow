import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTasks } from '../../hooks/useTasks';
import { ApiDataSourceImpl } from '../../data/datasources/ApiDataSourceImpl';

// Mock the API data source
jest.mock('../../data/datasources/ApiDataSourceImpl');

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useTasks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should fetch tasks for a repository', async () => {
    const mockTasks = [
      {
        id: '1',
        title: 'Test Task',
        description: 'Test Description',
        status: 'completed' as const,
        repository_id: 123,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ];

    const mockGetTasks = jest.fn().mockResolvedValue(mockTasks);
    (ApiDataSourceImpl as jest.MockedClass<typeof ApiDataSourceImpl>).prototype.getTasks = mockGetTasks;

    const { result } = renderHook(
      () => useTasks({ repositoryId: 123 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockGetTasks).toHaveBeenCalledWith(123);
    expect(result.current.data).toEqual(mockTasks);
  });

  it('should handle disabled state', () => {
    const { result } = renderHook(
      () => useTasks({ repositoryId: 123, enabled: false }),
      { wrapper: createWrapper() }
    );

    expect(result.current.isIdle).toBe(true);
  });

  it('should handle errors', async () => {
    const mockError = new Error('API Error');
    const mockGetTasks = jest.fn().mockRejectedValue(mockError);
    (ApiDataSourceImpl as jest.MockedClass<typeof ApiDataSourceImpl>).prototype.getTasks = mockGetTasks;

    const { result } = renderHook(
      () => useTasks({ repositoryId: 123 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toBe(mockError);
  });
});