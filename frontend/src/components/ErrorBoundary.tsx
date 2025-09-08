'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { ErrorHandler, ErrorType, AppError } from '../utils/errorHandler';
import { ErrorMessage } from './ErrorMessage';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: (error: AppError, resetError: () => void) => ReactNode;
  onError?: (error: AppError, errorInfo: ErrorInfo) => void;
  showDetails?: boolean;
  level?: 'page' | 'component' | 'critical';
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: AppError | null;
  errorId: string | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private resetTimeoutId: number | null = null;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorId: null,
    };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    const appError = ErrorHandler.createError(
      ErrorType.UNKNOWN,
      'COMPONENT_ERROR',
      `Component error: ${error.message}`,
      { originalError: error, recoverable: true }
    );

    const errorId = `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    return {
      hasError: true,
      error: appError,
      errorId,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const { error: appError } = this.state;
    
    if (appError) {
      // 상세 에러 정보 추가
      const enhancedError = {
        ...appError,
        message: `${appError.message}\nComponent Stack: ${errorInfo.componentStack}`,
        originalError: error,
      };

      // 에러 로깅
      ErrorHandler.logError(enhancedError, 'ErrorBoundary');
      
      // 사용자 정의 에러 핸들러 호출
      this.props.onError?.(enhancedError, errorInfo);

      // 개발 환경에서 상세 로깅
      if (process.env.NODE_ENV === 'development') {
        console.group('🚨 ErrorBoundary caught an error');
        console.error('Error:', error);
        console.error('Error Info:', errorInfo);
        console.error('App Error:', enhancedError);
        console.groupEnd();
      }
    }
  }

  resetError = () => {
    this.setState({
      hasError: false,
      error: null,
      errorId: null,
    });
  };

  handleRetry = async () => {
    this.resetError();
    
    // 자동 리셋 타이머 설정 (10초 후)
    this.resetTimeoutId = window.setTimeout(() => {
      if (this.state.hasError) {
        this.resetError();
      }
    }, 10000);
  };

  componentWillUnmount() {
    if (this.resetTimeoutId) {
      clearTimeout(this.resetTimeoutId);
    }
  }

  render() {
    const { hasError, error } = this.state;
    const { children, fallback, showDetails = false, level = 'component' } = this.props;

    if (hasError && error) {
      // 커스텀 fallback이 있다면 사용
      if (fallback) {
        return fallback(error, this.resetError);
      }

      // 레벨별 기본 UI
      switch (level) {
        case 'critical':
          return <CriticalErrorFallback error={error} onRetry={this.handleRetry} />;
        
        case 'page':
          return <PageErrorFallback error={error} onRetry={this.handleRetry} showDetails={showDetails} />;
        
        case 'component':
        default:
          return <ComponentErrorFallback error={error} onRetry={this.handleRetry} />;
      }
    }

    return children;
  }
}

// 컴포넌트 레벨 에러 UI
function ComponentErrorFallback({ 
  error, 
  onRetry 
}: { 
  error: AppError; 
  onRetry: () => void;
}) {
  return (
    <div className="p-4 border border-destructive/20 rounded-lg bg-destructive/5">
      <div className="flex items-center gap-2 text-destructive mb-2">
        <span className="text-sm font-medium">컴포넌트 오류</span>
      </div>
      <p className="text-sm text-muted-foreground mb-3">{error.userMessage}</p>
      <button
        onClick={onRetry}
        className="text-xs px-3 py-1.5 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors"
      >
        다시 시도
      </button>
    </div>
  );
}

// 페이지 레벨 에러 UI
function PageErrorFallback({ 
  error, 
  onRetry, 
  showDetails 
}: { 
  error: AppError; 
  onRetry: () => void;
  showDetails: boolean;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="w-full max-w-md">
        <ErrorMessage 
          error={error} 
          onRetry={onRetry}
          showDetails={showDetails}
        />
      </div>
    </div>
  );
}

// 크리티컬 에러 UI (전체 앱 실패)
function CriticalErrorFallback({ 
  error, 
  onRetry 
}: { 
  error: AppError; 
  onRetry: () => void;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-red-50">
      <div className="text-center space-y-6 max-w-md">
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-red-900">앱에 문제가 발생했습니다</h1>
          <p className="text-red-700">
            예상치 못한 오류가 발생했습니다. 잠시 후 다시 시도하거나 페이지를 새로고침 해주세요.
          </p>
        </div>
        
        <div className="space-y-3">
          <button
            onClick={onRetry}
            className="w-full px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            다시 시도
          </button>
          
          <button
            onClick={() => window.location.reload()}
            className="w-full px-4 py-2 border border-red-600 text-red-600 rounded-lg hover:bg-red-50 transition-colors"
          >
            페이지 새로고침
          </button>
          
          <button
            onClick={() => window.location.href = '/'}
            className="w-full px-4 py-2 text-red-600 hover:text-red-800 transition-colors"
          >
            홈으로 이동
          </button>
        </div>
        
        {process.env.NODE_ENV === 'development' && (
          <details className="text-left p-3 bg-red-100 rounded-lg text-xs">
            <summary className="cursor-pointer text-red-800 font-medium">기술적 정보</summary>
            <pre className="mt-2 text-red-700 overflow-auto">
              {JSON.stringify({
                type: error.type,
                code: error.code,
                message: error.message,
                statusCode: error.statusCode,
              }, null, 2)}
            </pre>
          </details>
        )}
      </div>
    </div>
  );
}

// Hook을 이용한 에러 바운더리 (함수 컴포넌트에서 사용)
export function useErrorBoundary() {
  return {
    captureError: (error: Error | AppError, context?: string) => {
      const appError = error instanceof Error ? ErrorHandler.fromError(error) : error;
      ErrorHandler.logError(appError, context);
      throw error; // 에러 바운더리가 캐치할 수 있도록 re-throw
    }
  };
}