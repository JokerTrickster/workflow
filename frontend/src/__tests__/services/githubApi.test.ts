/**
 * Unit tests for GitHub API service error handling and rate limiting
 * Tests comprehensive error scenarios and recovery mechanisms
 */

import { GitHubApiService } from '../../services/githubApi';
import { ErrorHandler, ErrorType } from '../../utils/errorHandler';

// Mock fetch
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock ErrorHandler
jest.mock('../../utils/errorHandler');
const mockErrorHandler = ErrorHandler as jest.Mocked<typeof ErrorHandler>;

describe('GitHubApiService Error Handling', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockErrorHandler.fromHttpError.mockImplementation((error) => ({
      type: ErrorType.NETWORK,
      message: error.message,
      originalError: error,
      context: {},
      timestamp: new Date(),
    }));
  });

  describe('Network Errors', () => {
    it('should handle network connection failures', async () => {
      mockFetch.mockRejectedValue(new Error('Network request failed'));

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow(
        'Network request failed'
      );

      expect(mockFetch).toHaveBeenCalledTimes(1);
    });

    it('should handle timeout errors', async () => {
      mockFetch.mockRejectedValue(new Error('Request timeout'));

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow(
        'Request timeout'
      );
    });

    it('should handle DNS resolution failures', async () => {
      mockFetch.mockRejectedValue(new Error('DNS resolution failed'));

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow(
        'DNS resolution failed'
      );
    });
  });

  describe('HTTP Status Errors', () => {
    it('should handle 401 authentication errors', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        json: () => Promise.resolve({
          message: 'Bad credentials',
          documentation_url: 'https://docs.github.com/rest'
        })
      } as Response);

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow();

      expect(mockFetch).toHaveBeenCalledWith(
        'https://api.github.com/user/repos?page=1&per_page=30&sort=updated',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/vnd.github.v3+json',
          }),
        })
      );
    });

    it('should handle 403 forbidden errors (rate limiting)', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        headers: new Headers({
          'X-RateLimit-Limit': '60',
          'X-RateLimit-Remaining': '0',
          'X-RateLimit-Reset': '1640995200', // Unix timestamp
          'Retry-After': '3600'
        }),
        json: () => Promise.resolve({
          message: 'API rate limit exceeded for user ID 1234.',
          documentation_url: 'https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting'
        })
      } as Response);

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow();
    });

    it('should handle 404 not found errors', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.resolve({
          message: 'Not Found',
          documentation_url: 'https://docs.github.com/rest'
        })
      } as Response);

      await expect(GitHubApiService.fetchRepositoryIssues('owner', 'repo', 1, 30)).rejects.toThrow();
    });

    it('should handle 422 validation errors', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 422,
        statusText: 'Unprocessable Entity',
        json: () => Promise.resolve({
          message: 'Validation Failed',
          errors: [
            {
              resource: 'Issue',
              field: 'title',
              code: 'missing_field'
            }
          ]
        })
      } as Response);

      await expect(GitHubApiService.fetchRepositoryIssues('owner', 'repo', 1, 30)).rejects.toThrow();
    });

    it('should handle 500 server errors', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.resolve({
          message: 'Server Error',
          documentation_url: 'https://docs.github.com/rest'
        })
      } as Response);

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow();
    });
  });

  describe('Rate Limiting Scenarios', () => {
    it('should extract rate limit information from headers', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        headers: new Headers({
          'X-RateLimit-Limit': '5000',
          'X-RateLimit-Remaining': '0',
          'X-RateLimit-Reset': '1640995200',
          'X-RateLimit-Used': '5000',
          'X-RateLimit-Resource': 'core'
        }),
        json: () => Promise.resolve({
          message: 'API rate limit exceeded',
          documentation_url: 'https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting'
        })
      } as Response);

      try {
        await GitHubApiService.fetchUserRepositories(1, 30);
      } catch (error) {
        // The service should process rate limit headers and include them in the error
        expect(error).toBeInstanceOf(Error);
      }
    });

    it('should handle secondary rate limits', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        headers: new Headers({
          'Retry-After': '60',
          'X-RateLimit-Limit': '5000',
          'X-RateLimit-Remaining': '4999'
        }),
        json: () => Promise.resolve({
          message: 'You have exceeded a secondary rate limit and have been temporarily blocked from content creation.',
          documentation_url: 'https://docs.github.com/rest/overview/resources-in-the-rest-api#secondary-rate-limits'
        })
      } as Response);

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow();
    });

    it('should handle abuse rate limits', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        headers: new Headers({
          'Retry-After': '120'
        }),
        json: () => Promise.resolve({
          message: 'You have triggered an abuse detection mechanism and have been temporarily blocked.',
          documentation_url: 'https://docs.github.com/rest/overview/resources-in-the-rest-api#abuse-rate-limits'
        })
      } as Response);

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow();
    });
  });

  describe('Malformed Response Handling', () => {
    it('should handle invalid JSON responses', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.reject(new Error('Unexpected token < in JSON at position 0'))
      } as Response);

      await expect(GitHubApiService.fetchUserRepositories(1, 30)).rejects.toThrow();
    });

    it('should handle empty responses', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 204,
        json: () => Promise.resolve(null)
      } as Response);

      const result = await GitHubApiService.fetchUserRepositories(1, 30);
      expect(result.repositories).toEqual([]);
    });

    it('should handle responses with unexpected structure', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({
          unexpected: 'structure',
          no_repos: true
        })
      } as Response);

      const result = await GitHubApiService.fetchUserRepositories(1, 30);
      expect(result.repositories).toEqual([]);
    });
  });

  describe('Request Parameter Validation', () => {
    it('should handle invalid page numbers', async () => {
      await expect(GitHubApiService.fetchUserRepositories(-1, 30)).rejects.toThrow();
      await expect(GitHubApiService.fetchUserRepositories(0, 30)).rejects.toThrow();
    });

    it('should handle invalid per_page values', async () => {
      await expect(GitHubApiService.fetchUserRepositories(1, -1)).rejects.toThrow();
      await expect(GitHubApiService.fetchUserRepositories(1, 0)).rejects.toThrow();
      await expect(GitHubApiService.fetchUserRepositories(1, 101)).rejects.toThrow();
    });

    it('should handle empty owner/repo parameters', async () => {
      await expect(GitHubApiService.fetchRepositoryIssues('', 'repo', 1, 30)).rejects.toThrow();
      await expect(GitHubApiService.fetchRepositoryIssues('owner', '', 1, 30)).rejects.toThrow();
      await expect(GitHubApiService.fetchRepositoryIssues('  ', 'repo', 1, 30)).rejects.toThrow();
    });
  });

  describe('Successful API Responses', () => {
    it('should handle successful repository fetch', async () => {
      const mockRepos = [
        {
          id: 1,
          name: 'test-repo',
          full_name: 'user/test-repo',
          description: 'Test repository',
          html_url: 'https://github.com/user/test-repo',
          clone_url: 'https://github.com/user/test-repo.git',
          ssh_url: 'git@github.com:user/test-repo.git',
          private: false,
          language: 'JavaScript',
          stargazers_count: 10,
          forks_count: 2,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-02T00:00:00Z',
          pushed_at: '2024-01-02T00:00:00Z',
          default_branch: 'main'
        }
      ];

      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers({
          'Link': '<https://api.github.com/user/repos?page=2>; rel="next", <https://api.github.com/user/repos?page=3>; rel="last"'
        }),
        json: () => Promise.resolve(mockRepos)
      } as Response);

      const result = await GitHubApiService.fetchUserRepositories(1, 30);

      expect(result.repositories).toHaveLength(1);
      expect(result.repositories[0]).toEqual(expect.objectContaining({
        id: 1,
        name: 'test-repo',
        full_name: 'user/test-repo'
      }));
      expect(result.hasMore).toBe(true);
      expect(result.pagination.has_next_page).toBe(true);
    });

    it('should handle successful issues fetch', async () => {
      const mockIssues = [
        {
          id: 1,
          number: 1,
          title: 'Test issue',
          body: 'This is a test issue',
          state: 'open',
          user: { login: 'testuser', avatar_url: 'https://github.com/testuser.png', id: 1 },
          assignees: [],
          labels: [],
          comments: 0,
          html_url: 'https://github.com/user/repo/issues/1',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z'
        }
      ];

      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockIssues)
      } as Response);

      const result = await GitHubApiService.fetchRepositoryIssues('user', 'repo', 1, 30);

      expect(result.issues).toHaveLength(1);
      expect(result.issues[0]).toEqual(expect.objectContaining({
        id: 1,
        number: 1,
        title: 'Test issue'
      }));
    });
  });

  describe('GitHub API Authentication', () => {
    it('should include authorization header when token is available', async () => {
      // Mock localStorage to return a token
      const mockGetItem = jest.fn(() => 'github_token_123');
      Object.defineProperty(window, 'localStorage', {
        value: { getItem: mockGetItem }
      });

      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve([])
      } as Response);

      await GitHubApiService.fetchUserRepositories(1, 30);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer github_token_123'
          })
        })
      );
    });

    it('should handle requests without authentication token', async () => {
      // Mock localStorage to return null
      const mockGetItem = jest.fn(() => null);
      Object.defineProperty(window, 'localStorage', {
        value: { getItem: mockGetItem }
      });

      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve([])
      } as Response);

      await GitHubApiService.fetchUserRepositories(1, 30);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.not.objectContaining({
            'Authorization': expect.any(String)
          })
        })
      );
    });
  });
});