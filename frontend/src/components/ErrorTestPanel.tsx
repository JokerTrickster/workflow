'use client';

import { useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { ErrorHandler, ErrorType } from '../utils/errorHandler';
import { ErrorMessage } from './ErrorMessage';
import { useErrorBoundary } from './ErrorBoundary';

export function ErrorTestPanel() {
  const [testError, setTestError] = useState<any>(null);
  const { captureError } = useErrorBoundary();

  // 개발 환경에서만 표시
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }

  const simulateNetworkError = () => {
    const error = ErrorHandler.createError(
      ErrorType.NETWORK,
      'NETWORK_TIMEOUT',
      'Simulated network timeout',
      { retryable: true }
    );
    setTestError(error);
  };

  const simulateAuthError = () => {
    const error = ErrorHandler.createError(
      ErrorType.AUTHENTICATION,
      'AUTH_UNAUTHORIZED',
      'Simulated auth error',
      { statusCode: 401, recoverable: true }
    );
    setTestError(error);
  };

  const simulateRateLimitError = () => {
    const error = ErrorHandler.createError(
      ErrorType.RATE_LIMIT,
      'RATE_LIMIT_GITHUB',
      'Simulated GitHub rate limit',
      { statusCode: 403, retryable: true }
    );
    setTestError(error);
  };

  const simulateServerError = () => {
    const error = ErrorHandler.createError(
      ErrorType.SERVER,
      'SERVER_ERROR',
      'Simulated server error',
      { statusCode: 500, retryable: true }
    );
    setTestError(error);
  };

  const simulateComponentError = () => {
    try {
      throw new Error('Simulated component error for ErrorBoundary test');
    } catch (error) {
      captureError(error as Error, 'ErrorTestPanel');
    }
  };

  const handleRetry = async () => {
    await new Promise(resolve => setTimeout(resolve, 1000)); // 1초 대기
    setTestError(null);
    alert('재시도가 성공했습니다!');
  };

  return (
    <div className="fixed bottom-4 left-4 z-50">
      <Card className="w-80">
        <CardHeader>
          <CardTitle className="text-sm">🧪 Error Testing Panel</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-2">
            <Button variant="outline" size="sm" onClick={simulateNetworkError}>
              Network Error
            </Button>
            <Button variant="outline" size="sm" onClick={simulateAuthError}>
              Auth Error
            </Button>
            <Button variant="outline" size="sm" onClick={simulateRateLimitError}>
              Rate Limit
            </Button>
            <Button variant="outline" size="sm" onClick={simulateServerError}>
              Server Error
            </Button>
          </div>
          
          <Button 
            variant="destructive" 
            size="sm" 
            className="w-full"
            onClick={simulateComponentError}
          >
            Component Error (ErrorBoundary Test)
          </Button>
          
          {testError && (
            <div className="mt-4">
              <h4 className="text-xs font-medium mb-2">Test Error Preview:</h4>
              <ErrorMessage
                error={testError}
                onRetry={handleRetry}
                onDismiss={() => setTestError(null)}
                compact={true}
              />
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}