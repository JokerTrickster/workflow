'use client';

import { useState, useCallback } from 'react';
import { AppError, ErrorHandler, ErrorType } from '../utils/errorHandler';

interface ErrorRecoveryOptions {
  maxRetries?: number;
  retryDelay?: number;
  onRetrySuccess?: () => void;
  onRetryFailure?: (error: AppError) => void;
  onMaxRetriesReached?: (error: AppError) => void;
}

interface ErrorRecoveryState {
  error: AppError | null;
  isRetrying: boolean;
  retryCount: number;
  canRetry: boolean;
  canRecover: boolean;
}

export function useErrorRecovery(options: ErrorRecoveryOptions = {}) {
  const [state, setState] = useState<ErrorRecoveryState>({
    error: null,
    isRetrying: false,
    retryCount: 0,
    canRetry: false,
    canRecover: false,
  });

  const setError = useCallback((error: AppError | Error | null, context?: string) => {
    if (!error) {
      setState({
        error: null,
        isRetrying: false,
        retryCount: 0,
        canRetry: false,
        canRecover: false,
      });
      return;
    }

    const appError = error instanceof Error ? ErrorHandler.fromError(error) : error;
    
    // 에러 로깅
    ErrorHandler.logError(appError, context);

    setState(prev => ({
      error: appError,
      isRetrying: false,
      retryCount: appError === prev.error ? prev.retryCount : 0,
      canRetry: ErrorHandler.shouldRetry(appError, prev.retryCount),
      canRecover: appError.recoverable,
    }));
  }, []);

  const retry = useCallback(async (retryFn?: () => Promise<any>) => {
    const { error, retryCount } = state;
    
    if (!error || !state.canRetry) {
      return;
    }

    setState(prev => ({ ...prev, isRetrying: true }));

    const delay = options.retryDelay ?? ErrorHandler.getRetryDelay(error, retryCount);
    
    try {
      await new Promise(resolve => setTimeout(resolve, delay));
      
      if (retryFn) {
        await retryFn();
        
        // 성공 시 에러 상태 초기화
        setState({
          error: null,
          isRetrying: false,
          retryCount: 0,
          canRetry: false,
          canRecover: false,
        });
        
        options.onRetrySuccess?.();
      }
    } catch (retryError) {
      const newRetryCount = retryCount + 1;
      const appError = retryError instanceof Error ? ErrorHandler.fromError(retryError) : retryError as AppError;
      
      const canRetryAgain = ErrorHandler.shouldRetry(appError, newRetryCount);
      
      setState({
        error: appError,
        isRetrying: false,
        retryCount: newRetryCount,
        canRetry: canRetryAgain,
        canRecover: appError.recoverable,
      });

      if (canRetryAgain) {
        options.onRetryFailure?.(appError);
      } else {
        options.onMaxRetriesReached?.(appError);
      }
    }
  }, [state, options]);

  const recover = useCallback(() => {
    const { error } = state;
    
    if (!error || !state.canRecover) {
      return;
    }

    // 에러 타입별 복구 액션
    switch (error.type) {
      case ErrorType.AUTHENTICATION:
        // 로그인 페이지로 리디렉션
        window.location.href = '/login';
        break;
        
      case ErrorType.AUTHORIZATION:
        // GitHub 권한 재설정 페이지로 이동
        if (error.code === 'AUTH_INSUFFICIENT_SCOPE') {
          window.location.href = '/auth/permissions';
        }
        break;
        
      case ErrorType.NETWORK:
        // 온라인 상태 확인 후 새로고침
        if (navigator.onLine) {
          window.location.reload();
        }
        break;
        
      case ErrorType.RATE_LIMIT:
        // Rate limit 정보 표시 및 대기
        if (error.code === 'RATE_LIMIT_GITHUB') {
          alert('GitHub API 사용량을 초과했습니다. 1시간 후에 다시 시도해주세요.');
        }
        break;
        
      case ErrorType.SERVER:
        // 홈페이지로 이동
        window.location.href = '/';
        break;
        
      default:
        // 페이지 새로고침
        window.location.reload();
    }
  }, [state]);

  const clearError = useCallback(() => {
    setState({
      error: null,
      isRetrying: false,
      retryCount: 0,
      canRetry: false,
      canRecover: false,
    });
  }, []);

  const getRecoveryAction = useCallback((error: AppError): {
    label: string;
    action: () => void;
  } | null => {
    if (!error.recoverable) return null;

    switch (error.type) {
      case ErrorType.AUTHENTICATION:
        return {
          label: '다시 로그인',
          action: () => window.location.href = '/login'
        };
        
      case ErrorType.AUTHORIZATION:
        if (error.code === 'AUTH_INSUFFICIENT_SCOPE') {
          return {
            label: '권한 재설정',
            action: () => window.location.href = '/auth/permissions'
          };
        }
        return {
          label: '홈으로 이동',
          action: () => window.location.href = '/'
        };
        
      case ErrorType.NETWORK:
        return {
          label: '새로고침',
          action: () => window.location.reload()
        };
        
      case ErrorType.SERVER:
        return {
          label: '홈으로 이동',
          action: () => window.location.href = '/'
        };
        
      case ErrorType.VALIDATION:
        return {
          label: '다시 시도',
          action: clearError
        };
        
      default:
        return {
          label: '새로고침',
          action: () => window.location.reload()
        };
    }
  }, [clearError]);

  return {
    ...state,
    setError,
    retry,
    recover,
    clearError,
    getRecoveryAction,
  };
}