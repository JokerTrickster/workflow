export enum ErrorType {
  NETWORK = 'NETWORK',
  AUTHENTICATION = 'AUTHENTICATION',
  AUTHORIZATION = 'AUTHORIZATION',
  RATE_LIMIT = 'RATE_LIMIT',
  SERVER = 'SERVER',
  VALIDATION = 'VALIDATION',
  UNKNOWN = 'UNKNOWN',
}

export interface AppError {
  type: ErrorType;
  code: string;
  message: string;
  userMessage: string;
  recoverable: boolean;
  retryable: boolean;
  statusCode?: number;
  originalError?: Error;
}

export class ErrorHandler {
  private static errorMessages: Record<string, string> = {
    // 네트워크 에러
    'NETWORK_OFFLINE': '인터넷 연결을 확인해주세요.',
    'NETWORK_TIMEOUT': '서버 응답이 지연되고 있습니다. 잠시 후 다시 시도해주세요.',
    'NETWORK_UNKNOWN': '네트워크 오류가 발생했습니다. 연결을 확인해주세요.',
    
    // 인증 에러
    'AUTH_UNAUTHORIZED': '로그인이 필요합니다. 다시 로그인해주세요.',
    'AUTH_TOKEN_EXPIRED': '로그인 세션이 만료되었습니다. 다시 로그인해주세요.',
    'AUTH_INVALID_CREDENTIALS': '잘못된 인증 정보입니다.',
    
    // 권한 에러
    'AUTH_FORBIDDEN': '이 작업을 수행할 권한이 없습니다.',
    'AUTH_INSUFFICIENT_SCOPE': 'GitHub 권한이 부족합니다. 권한을 다시 설정해주세요.',
    
    // Rate Limit 에러
    'RATE_LIMIT_EXCEEDED': 'API 호출 한도를 초과했습니다. 잠시 후 다시 시도해주세요.',
    'RATE_LIMIT_GITHUB': 'GitHub API 한도를 초과했습니다. 1시간 후 다시 시도해주세요.',
    
    // 서버 에러
    'SERVER_ERROR': '서버에 문제가 발생했습니다. 잠시 후 다시 시도해주세요.',
    'SERVER_MAINTENANCE': '서버가 점검 중입니다. 잠시 후 다시 시도해주세요.',
    'SERVER_UNAVAILABLE': '서버를 사용할 수 없습니다. 잠시 후 다시 시도해주세요.',
    
    // 검증 에러
    'VALIDATION_REQUIRED': '필수 정보가 누락되었습니다.',
    'VALIDATION_INVALID_FORMAT': '입력 형식이 올바르지 않습니다.',
    'VALIDATION_TOO_LONG': '입력 내용이 너무 깁니다.',
    
    // 일반 에러
    'UNKNOWN_ERROR': '알 수 없는 오류가 발생했습니다. 계속 문제가 발생하면 관리자에게 문의해주세요.',
  };

  static createError(
    type: ErrorType,
    code: string,
    message: string,
    options: {
      statusCode?: number;
      originalError?: Error;
      recoverable?: boolean;
      retryable?: boolean;
    } = {}
  ): AppError {
    const userMessage = this.errorMessages[code] || this.errorMessages['UNKNOWN_ERROR'];
    
    return {
      type,
      code,
      message,
      userMessage,
      recoverable: options.recoverable ?? this.isRecoverable(type, code),
      retryable: options.retryable ?? this.isRetryable(type, code),
      statusCode: options.statusCode,
      originalError: options.originalError,
    };
  }

  static fromHttpError(error: any): AppError {
    const status = error.status || error.response?.status;
    const statusText = error.statusText || error.response?.statusText;
    const data = error.response?.data || error.data || {};

    switch (status) {
      case 401:
        return this.createError(
          ErrorType.AUTHENTICATION,
          'AUTH_UNAUTHORIZED',
          'Unauthorized access',
          { statusCode: status, originalError: error }
        );

      case 403:
        if (data.message?.includes('rate limit') || data.message?.includes('API rate limit')) {
          return this.createError(
            ErrorType.RATE_LIMIT,
            status === 403 && data.message?.includes('GitHub') ? 'RATE_LIMIT_GITHUB' : 'RATE_LIMIT_EXCEEDED',
            'Rate limit exceeded',
            { statusCode: status, originalError: error }
          );
        }
        return this.createError(
          ErrorType.AUTHORIZATION,
          'AUTH_FORBIDDEN',
          'Forbidden access',
          { statusCode: status, originalError: error }
        );

      case 422:
        return this.createError(
          ErrorType.VALIDATION,
          'VALIDATION_INVALID_FORMAT',
          'Validation error',
          { statusCode: status, originalError: error }
        );

      case 429:
        return this.createError(
          ErrorType.RATE_LIMIT,
          'RATE_LIMIT_EXCEEDED',
          'Too many requests',
          { statusCode: status, originalError: error }
        );

      case 500:
      case 502:
      case 503:
      case 504:
        return this.createError(
          ErrorType.SERVER,
          status === 503 ? 'SERVER_UNAVAILABLE' : 'SERVER_ERROR',
          `Server error: ${statusText}`,
          { statusCode: status, originalError: error }
        );

      default:
        if (!status) {
          // Network error
          if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
            return this.createError(
              ErrorType.NETWORK,
              'NETWORK_TIMEOUT',
              'Request timeout',
              { originalError: error }
            );
          }
          
          if (error.code === 'ERR_NETWORK' || !navigator.onLine) {
            return this.createError(
              ErrorType.NETWORK,
              'NETWORK_OFFLINE',
              'Network error',
              { originalError: error }
            );
          }

          return this.createError(
            ErrorType.NETWORK,
            'NETWORK_UNKNOWN',
            'Network error',
            { originalError: error }
          );
        }

        return this.createError(
          ErrorType.UNKNOWN,
          'UNKNOWN_ERROR',
          `Unknown error: ${statusText || 'Unknown'}`,
          { statusCode: status, originalError: error }
        );
    }
  }

  static fromError(error: Error): AppError {
    if ('status' in error || 'response' in error) {
      return this.fromHttpError(error);
    }

    return this.createError(
      ErrorType.UNKNOWN,
      'UNKNOWN_ERROR',
      error.message,
      { originalError: error }
    );
  }

  private static isRecoverable(type: ErrorType, code: string): boolean {
    switch (type) {
      case ErrorType.AUTHENTICATION:
      case ErrorType.AUTHORIZATION:
        return true; // 로그인/권한 재설정 가능
      case ErrorType.NETWORK:
        return true; // 네트워크 복구 후 재시도 가능
      case ErrorType.RATE_LIMIT:
        return true; // 시간 지난 후 재시도 가능
      case ErrorType.VALIDATION:
        return true; // 사용자가 입력 수정 가능
      case ErrorType.SERVER:
        return code !== 'SERVER_MAINTENANCE'; // 점검 중이 아니면 복구 가능
      default:
        return false;
    }
  }

  private static isRetryable(type: ErrorType, code: string): boolean {
    switch (type) {
      case ErrorType.NETWORK:
        return true;
      case ErrorType.SERVER:
        return code !== 'SERVER_MAINTENANCE';
      case ErrorType.RATE_LIMIT:
        return true; // 시간 지난 후
      default:
        return false;
    }
  }

  static getRetryDelay(error: AppError, attemptCount: number): number {
    switch (error.type) {
      case ErrorType.RATE_LIMIT:
        if (error.code === 'RATE_LIMIT_GITHUB') {
          return 60 * 60 * 1000; // 1시간
        }
        return Math.min(1000 * Math.pow(2, attemptCount), 30000); // 지수 백오프, 최대 30초
      
      case ErrorType.NETWORK:
      case ErrorType.SERVER:
        return Math.min(1000 * Math.pow(2, attemptCount), 10000); // 지수 백오프, 최대 10초
      
      default:
        return 1000;
    }
  }

  static shouldRetry(error: AppError, attemptCount: number): boolean {
    if (!error.retryable) return false;
    
    switch (error.type) {
      case ErrorType.RATE_LIMIT:
        return attemptCount < 2; // 최대 2회 재시도
      case ErrorType.NETWORK:
      case ErrorType.SERVER:
        return attemptCount < 3; // 최대 3회 재시도
      default:
        return false;
    }
  }

  static logError(error: AppError, context?: string): void {
    const logData = {
      type: error.type,
      code: error.code,
      message: error.message,
      statusCode: error.statusCode,
      context,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href,
    };

    // 개발 환경에서는 콘솔에 출력
    if (process.env.NODE_ENV === 'development') {
      console.group(`🚨 Error [${error.type}]`);
      console.error('Error Details:', logData);
      console.error('Original Error:', error.originalError);
      console.groupEnd();
    }

    // 프로덕션에서는 에러 추적 서비스로 전송
    // TODO: Sentry, LogRocket 등의 에러 추적 서비스 연동
  }
}