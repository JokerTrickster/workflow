import { renderHook, act } from '@testing-library/react'
import { useErrorRecovery } from '../../hooks/useErrorRecovery'
import { AppError, ErrorType } from '../../utils/errorHandler'

const mockError: AppError = {
  id: 'test-error',
  type: ErrorType.NETWORK,
  message: 'Network error',
  userMessage: '네트워크 오류가 발생했습니다',
  timestamp: new Date(),
  context: 'Test context',
  retry: {
    canRetry: true,
    attempts: 1,
    maxAttempts: 3,
    nextRetryAt: new Date(Date.now() + 5000)
  }
}

describe('useErrorRecovery Hook', () => {
  beforeEach(() => {
    jest.clearAllTimers()
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.runOnlyPendingTimers()
    jest.useRealTimers()
  })

  it('initializes with no error', () => {
    const { result } = renderHook(() => useErrorRecovery())
    
    expect(result.current.error).toBeNull()
    expect(result.current.isRetrying).toBe(false)
    expect(result.current.retryCount).toBe(0)
  })

  it('sets error when setError is called', () => {
    const { result } = renderHook(() => useErrorRecovery())
    
    act(() => {
      result.current.setError(mockError, 'Test component')
    })
    
    expect(result.current.error).toEqual(mockError)
    expect(result.current.isRetrying).toBe(false)
  })

  it('clears error when clearError is called', () => {
    const { result } = renderHook(() => useErrorRecovery())
    
    act(() => {
      result.current.setError(mockError, 'Test component')
    })
    
    expect(result.current.error).toEqual(mockError)
    
    act(() => {
      result.current.clearError()
    })
    
    expect(result.current.error).toBeNull()
  })

  it('increments retry count when retry is called', () => {
    const { result } = renderHook(() => useErrorRecovery())
    
    act(() => {
      result.current.setError(mockError, 'Test component')
    })
    
    act(() => {
      result.current.retry()
    })
    
    expect(result.current.retryCount).toBe(1)
    expect(result.current.isRetrying).toBe(true)
  })

  it('automatically retries when auto-retry is enabled', () => {
    const autoRetryError = {
      ...mockError,
      retry: {
        ...mockError.retry,
        nextRetryAt: new Date(Date.now() + 1000)
      }
    }
    
    const { result } = renderHook(() => useErrorRecovery({ autoRetry: true }))
    
    act(() => {
      result.current.setError(autoRetryError, 'Test component')
    })
    
    expect(result.current.error).toEqual(autoRetryError)
    
    // Fast-forward time to trigger auto-retry
    act(() => {
      jest.advanceTimersByTime(1000)
    })
    
    expect(result.current.retryCount).toBe(1)
  })

  it('does not auto-retry when max attempts reached', () => {
    const maxAttemptsError = {
      ...mockError,
      retry: {
        ...mockError.retry,
        attempts: 3,
        maxAttempts: 3
      }
    }
    
    const { result } = renderHook(() => useErrorRecovery({ autoRetry: true }))
    
    act(() => {
      result.current.setError(maxAttemptsError, 'Test component')
    })
    
    // Fast-forward time
    act(() => {
      jest.advanceTimersByTime(5000)
    })
    
    expect(result.current.retryCount).toBe(0) // Should not retry
  })

  it('calls onRetry callback when retry is triggered', () => {
    const onRetryMock = jest.fn()
    const { result } = renderHook(() => useErrorRecovery({ onRetry: onRetryMock }))
    
    act(() => {
      result.current.setError(mockError, 'Test component')
    })
    
    act(() => {
      result.current.retry()
    })
    
    expect(onRetryMock).toHaveBeenCalledWith(mockError, 'Test component')
  })

  it('handles multiple errors correctly', () => {
    const { result } = renderHook(() => useErrorRecovery())
    
    const firstError = { ...mockError, id: 'error-1' }
    const secondError = { ...mockError, id: 'error-2' }
    
    act(() => {
      result.current.setError(firstError, 'Component 1')
    })
    
    expect(result.current.error).toEqual(firstError)
    
    act(() => {
      result.current.setError(secondError, 'Component 2')
    })
    
    expect(result.current.error).toEqual(secondError) // Should replace with latest error
  })

  it('resets retry state when new error is set', () => {
    const { result } = renderHook(() => useErrorRecovery())
    
    act(() => {
      result.current.setError(mockError, 'Test component')
    })
    
    act(() => {
      result.current.retry()
    })
    
    expect(result.current.retryCount).toBe(1)
    
    const newError = { ...mockError, id: 'new-error' }
    
    act(() => {
      result.current.setError(newError, 'Test component')
    })
    
    expect(result.current.retryCount).toBe(0) // Should reset
    expect(result.current.isRetrying).toBe(false)
  })
})