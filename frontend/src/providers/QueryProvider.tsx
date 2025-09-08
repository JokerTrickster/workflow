'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import { ErrorHandler } from '../utils/errorHandler';

export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000, // 1 minute
            refetchOnWindowFocus: false,
            retry: (failureCount, error) => {
              const appError = ErrorHandler.fromHttpError(error);
              return ErrorHandler.shouldRetry(appError, failureCount);
            },
            retryDelay: (attemptIndex, error) => {
              const appError = ErrorHandler.fromHttpError(error);
              return ErrorHandler.getRetryDelay(appError, attemptIndex);
            },
          },
          mutations: {
            retry: (failureCount, error) => {
              const appError = ErrorHandler.fromHttpError(error);
              return ErrorHandler.shouldRetry(appError, failureCount);
            },
            retryDelay: (attemptIndex, error) => {
              const appError = ErrorHandler.fromHttpError(error);
              return ErrorHandler.getRetryDelay(appError, attemptIndex);
            },
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}