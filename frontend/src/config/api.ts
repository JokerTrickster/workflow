/**
 * API Configuration
 * Central place to manage all API endpoints and configuration
 */

interface ApiConfig {
  baseUrl: string;
  endpoints: {
    tasks: {
      list: (repositoryName?: string) => string;
      create: () => string;
      get: (id: string) => string;
      update: (id: string) => string;
      delete: (id: string) => string;
      execute: (id: string) => string;
    };
    auth: {
      login: () => string;
      logout: () => string;
      me: () => string;
      github: () => string;
    };
    repos: {
      list: () => string;
      clone: () => string;
      status: (id: number) => string;
    };
    workLogs: {
      list: () => string;
      entry: () => string;
    };
    github: {
      pulls: (owner: string, repo: string, number: number) => string;
      merge: (owner: string, repo: string, number: number) => string;
    };
  };
}

/**
 * Get API base URL from environment variables
 */
const getApiBaseUrl = (): string => {
  // Always use direct backend URL (7000 port)
  return process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:7000/api/v1';
};

/**
 * Create API configuration
 */
const createApiConfig = (): ApiConfig => {
  const baseUrl = getApiBaseUrl();

  return {
    baseUrl,
    endpoints: {
      tasks: {
        list: (repositoryName?: string) => {
          const params = repositoryName ? `?repository_name=${encodeURIComponent(repositoryName)}` : '';
          return `${baseUrl}/tasks${params}`;
        },
        create: () => `${baseUrl}/tasks`,
        get: (id: string) => `${baseUrl}/tasks/${id}`,
        update: (id: string) => `${baseUrl}/tasks/${id}`,
        delete: (id: string) => `${baseUrl}/tasks/${id}`,
        execute: (id: string) => `${baseUrl}/tasks/${id}/execute`,
      },
      auth: {
        login: () => '/api/auth/github',
        logout: () => `${baseUrl}/auth/logout`,
        me: () => `${baseUrl}/auth/me`,
        github: () => '/api/auth/github',
      },
      repos: {
        list: () => `${baseUrl}/repos`,
        clone: () => `${baseUrl}/repos/clone`,
        status: (id: number) => `${baseUrl}/repos/${id}/status`,
      },
      workLogs: {
        list: () => `${baseUrl}/work-logs`,
        entry: () => `${baseUrl}/work-logs/entry`,
      },
      github: {
        pulls: (owner: string, repo: string, number: number) =>
          `${baseUrl}/github/repos/${owner}/${repo}/pulls/${number}`,
        merge: (owner: string, repo: string, number: number) =>
          `${baseUrl}/github/repos/${owner}/${repo}/pulls/${number}/merge`,
      },
    },
  };
};

/**
 * API configuration instance
 */
export const apiConfig = createApiConfig();

/**
 * Helper function to build full API URLs
 */
export const buildApiUrl = (endpoint: string, params?: Record<string, string>): string => {
  let url = endpoint;

  if (params) {
    const searchParams = new URLSearchParams(params);
    const paramString = searchParams.toString();
    if (paramString) {
      url += url.includes('?') ? `&${paramString}` : `?${paramString}`;
    }
  }

  return url;
};

/**
 * Get API base URL (exported for compatibility)
 */
export { getApiBaseUrl };

/**
 * Environment-specific configuration
 */
export const config = {
  api: apiConfig,
  isDevelopment: process.env.NODE_ENV === 'development',
  isProduction: process.env.NODE_ENV === 'production',
  isTest: process.env.NODE_ENV === 'test',
} as const;