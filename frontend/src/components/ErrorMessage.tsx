'use client';

import { AlertTriangle, RefreshCw, WifiOff, Clock, Shield, X } from 'lucide-react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader } from './ui/card';
import { AppError, ErrorType } from '../utils/errorHandler';
import { useErrorRecovery } from '../hooks/useErrorRecovery';

interface ErrorMessageProps {
  error: AppError;
  onRetry?: () => Promise<any>;
  onDismiss?: () => void;
  showDetails?: boolean;
  compact?: boolean;
  className?: string;
}

const ErrorIcons = {
  [ErrorType.NETWORK]: WifiOff,
  [ErrorType.AUTHENTICATION]: Shield,
  [ErrorType.AUTHORIZATION]: Shield,
  [ErrorType.RATE_LIMIT]: Clock,
  [ErrorType.SERVER]: AlertTriangle,
  [ErrorType.VALIDATION]: AlertTriangle,
  [ErrorType.UNKNOWN]: AlertTriangle,
};

const ErrorColors = {
  [ErrorType.NETWORK]: 'text-blue-600 bg-blue-100',
  [ErrorType.AUTHENTICATION]: 'text-orange-600 bg-orange-100',
  [ErrorType.AUTHORIZATION]: 'text-red-600 bg-red-100',
  [ErrorType.RATE_LIMIT]: 'text-yellow-600 bg-yellow-100',
  [ErrorType.SERVER]: 'text-red-600 bg-red-100',
  [ErrorType.VALIDATION]: 'text-yellow-600 bg-yellow-100',
  [ErrorType.UNKNOWN]: 'text-gray-600 bg-gray-100',
};

export function ErrorMessage({
  error,
  onRetry,
  onDismiss,
  showDetails = false,
  compact = false,
  className = '',
}: ErrorMessageProps) {
  const {
    isRetrying,
    canRetry,
    canRecover,
    retry,
    getRecoveryAction
  } = useErrorRecovery({
    onRetrySuccess: onDismiss,
  });

  const IconComponent = ErrorIcons[error.type];
  const colorClass = ErrorColors[error.type];
  const recoveryAction = getRecoveryAction(error);

  const handleRetry = async () => {
    if (onRetry) {
      await retry(onRetry);
    } else {
      window.location.reload();
    }
  };

  if (compact) {
    return (
      <div className={`flex items-center gap-3 p-3 rounded-lg border bg-destructive/10 border-destructive/20 ${className}`}>
        <IconComponent className="h-4 w-4 text-destructive flex-shrink-0" />
        <span className="text-sm text-destructive flex-1">{error.userMessage}</span>
        
        <div className="flex items-center gap-2">
          {(canRetry || error.retryable) && onRetry && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRetry}
              disabled={isRetrying}
              className="h-7 px-2 text-destructive hover:bg-destructive/20"
            >
              <RefreshCw className={`h-3 w-3 ${isRetrying ? 'animate-spin' : ''}`} />
            </Button>
          )}
          
          {onDismiss && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onDismiss}
              className="h-7 w-7 p-0 text-destructive hover:bg-destructive/20"
            >
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>
      </div>
    );
  }

  return (
    <Card className={`border-destructive/20 ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-start gap-3">
          <div className={`p-2 rounded-full ${colorClass}`}>
            <IconComponent className="h-5 w-5" />
          </div>
          
          <div className="flex-1">
            <h3 className="font-semibold text-foreground">
              {getErrorTitle(error.type)}
            </h3>
            <p className="text-sm text-muted-foreground mt-1">
              {error.userMessage}
            </p>
          </div>
          
          {onDismiss && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onDismiss}
              className="h-8 w-8 p-0"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </CardHeader>
      
      {(showDetails || (canRetry && onRetry) || canRecover) && (
        <CardContent className="pt-0">
          {showDetails && (
            <div className="mb-4 p-3 rounded-md bg-muted">
              <p className="text-xs text-muted-foreground mb-1">기술적 정보:</p>
              <p className="text-xs font-mono text-muted-foreground">
                {error.code} {error.statusCode && `(${error.statusCode})`}
              </p>
              {error.message && (
                <p className="text-xs font-mono text-muted-foreground mt-1">
                  {error.message}
                </p>
              )}
            </div>
          )}
          
          <div className="flex gap-2 flex-wrap">
            {(canRetry || error.retryable) && onRetry && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleRetry}
                disabled={isRetrying}
                className="gap-2"
              >
                <RefreshCw className={`h-4 w-4 ${isRetrying ? 'animate-spin' : ''}`} />
                {isRetrying ? '다시 시도 중...' : '다시 시도'}
              </Button>
            )}
            
            {canRecover && recoveryAction && (
              <Button
                variant="default"
                size="sm"
                onClick={recoveryAction.action}
                className="gap-2"
              >
                {recoveryAction.label}
              </Button>
            )}
            
            {error.type === ErrorType.NETWORK && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => window.location.reload()}
                className="gap-2"
              >
                <RefreshCw className="h-4 w-4" />
                새로고침
              </Button>
            )}
            
            {error.type === ErrorType.RATE_LIMIT && error.code === 'RATE_LIMIT_GITHUB' && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Clock className="h-4 w-4" />
                1시간 후 다시 시도 가능
              </div>
            )}
          </div>
        </CardContent>
      )}
    </Card>
  );
}

function getErrorTitle(type: ErrorType): string {
  switch (type) {
    case ErrorType.NETWORK:
      return '연결 문제';
    case ErrorType.AUTHENTICATION:
      return '로그인 필요';
    case ErrorType.AUTHORIZATION:
      return '권한 부족';
    case ErrorType.RATE_LIMIT:
      return '사용량 초과';
    case ErrorType.SERVER:
      return '서버 오류';
    case ErrorType.VALIDATION:
      return '입력 오류';
    default:
      return '오류 발생';
  }
}

// 토스트/스낵바 스타일의 간단한 에러 메시지
export function ErrorToast({
  error,
  onRetry,
  onDismiss,
}: {
  error: AppError;
  onRetry?: () => Promise<any>;
  onDismiss?: () => void;
}) {
  return (
    <div className="fixed bottom-4 right-4 z-50 max-w-md">
      <ErrorMessage
        error={error}
        onRetry={onRetry}
        onDismiss={onDismiss}
        compact={true}
        className="shadow-lg"
      />
    </div>
  );
}

// 전체 페이지를 가리는 에러 오버레이
export function ErrorOverlay({
  error,
  onRetry,
  showDetails = true,
}: {
  error: AppError;
  onRetry?: () => Promise<any>;
  showDetails?: boolean;
}) {
  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <ErrorMessage
          error={error}
          onRetry={onRetry}
          showDetails={showDetails}
          className="shadow-2xl"
        />
      </div>
    </div>
  );
}