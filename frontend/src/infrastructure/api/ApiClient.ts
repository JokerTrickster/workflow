import { ErrorHandler, ErrorType } from '../../utils/errorHandler';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';

interface ApiResponse<T = any> {
  data: T;
  message?: string;
  success: boolean;
}

class ApiClient {
  private baseURL: string;

  constructor() {
    this.baseURL = API_BASE_URL;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const token = this.getAuthToken();

    const defaultHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (token) {
      defaultHeaders.Authorization = `Bearer ${token}`;
    }

    const config: RequestInit = {
      timeout: 10000, // 10초 타임아웃
      ...options,
      headers: {
        ...defaultHeaders,
        ...options.headers,
      },
    };

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), config.timeout as number);

      const response = await fetch(url, {
        ...config,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      // 상태 코드별 에러 처리
      if (!response.ok) {
        const errorData = await this.parseErrorResponse(response);
        throw this.createHttpError(response, errorData);
      }

      // 응답 파싱
      const data = await response.json();
      
      // API 응답 형식 검증
      if (data && typeof data === 'object' && 'success' in data) {
        if (!data.success) {
          throw new Error(data.message || 'API request failed');
        }
        return data.data;
      }

      return data;

    } catch (error) {
      // AbortError (타임아웃) 처리
      if (error instanceof Error && error.name === 'AbortError') {
        throw ErrorHandler.createError(
          ErrorType.NETWORK,
          'NETWORK_TIMEOUT',
          'Request timeout',
          { originalError: error, retryable: true }
        );
      }

      // 네트워크 에러 처리
      if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
        throw ErrorHandler.createError(
          ErrorType.NETWORK,
          'NETWORK_OFFLINE',
          'Network connection failed',
          { originalError: error, recoverable: true }
        );
      }

      // 이미 처리된 HTTP 에러면 그대로 throw
      if (error && typeof error === 'object' && 'type' in error) {
        throw error;
      }

      // 알 수 없는 에러
      throw ErrorHandler.fromError(error as Error);
    }
  }

  private getAuthToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
  }

  private async parseErrorResponse(response: Response): Promise<any> {
    try {
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      }
      return { message: await response.text() };
    } catch {
      return { message: response.statusText };
    }
  }

  private createHttpError(response: Response, errorData: any): Error {
    const error = new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
    
    // HTTP 응답 정보 추가
    (error as any).status = response.status;
    (error as any).statusText = response.statusText;
    (error as any).response = {
      status: response.status,
      statusText: response.statusText,
      data: errorData,
    };

    return error;
  }

  // 토큰 만료 체크 및 자동 갱신
  private async handleTokenExpiry(): Promise<void> {
    // 토큰 제거
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
      sessionStorage.removeItem('auth_token');
    }

    // 로그인 페이지로 리디렉션 (현재 페이지 정보 저장)
    const currentPath = window.location.pathname;
    sessionStorage.setItem('redirect_after_login', currentPath);
    window.location.href = '/login';
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data?: Record<string, unknown>): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: Record<string, unknown>): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

export const apiClient = new ApiClient();