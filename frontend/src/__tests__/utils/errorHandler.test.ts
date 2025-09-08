import { ErrorHandler, ErrorType, AppError } from '../../utils/errorHandler'

describe('ErrorHandler', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('fromHttpError', () => {
    it('classifies 401 errors as AUTHENTICATION', () => {
      const httpError = new Error('Unauthorized') as any
      httpError.status = 401
      
      const appError = ErrorHandler.fromHttpError(httpError)
      
      expect(appError.type).toBe(ErrorType.AUTHENTICATION)
      expect(appError.userMessage).toBe('로그인이 필요합니다')
    })

    it('classifies 403 errors as AUTHORIZATION', () => {
      const httpError = new Error('Forbidden') as any
      httpError.status = 403
      
      const appError = ErrorHandler.fromHttpError(httpError)
      
      expect(appError.type).toBe(ErrorType.AUTHORIZATION)
      expect(appError.userMessage).toBe('접근 권한이 없습니다')
    })

    it('classifies 429 errors as RATE_LIMIT', () => {
      const httpError = new Error('Too Many Requests') as any
      httpError.status = 429
      
      const appError = ErrorHandler.fromHttpError(httpError)
      
      expect(appError.type).toBe(ErrorType.RATE_LIMIT)
      expect(appError.userMessage).toBe('요청이 너무 많습니다. 잠시 후 다시 시도해주세요')
    })

    it('classifies network errors as NETWORK', () => {
      const networkError = new Error('Network Error')
      
      const appError = ErrorHandler.fromHttpError(networkError)
      
      expect(appError.type).toBe(ErrorType.NETWORK)
      expect(appError.userMessage).toBe('네트워크 연결을 확인해주세요')
    })

    it('sets retry information for retryable errors', () => {
      const httpError = new Error('Service Unavailable') as any
      httpError.status = 503
      
      const appError = ErrorHandler.fromHttpError(httpError)
      
      expect(appError.retry.canRetry).toBe(true)
      expect(appError.retry.attempts).toBe(0)
      expect(appError.retry.maxAttempts).toBe(3)
      expect(appError.retry.nextRetryAt).toBeInstanceOf(Date)
    })

    it('sets no retry for non-retryable errors', () => {
      const httpError = new Error('Bad Request') as any
      httpError.status = 400
      
      const appError = ErrorHandler.fromHttpError(httpError)
      
      expect(appError.retry.canRetry).toBe(false)
      expect(appError.retry.maxAttempts).toBe(0)
    })
  })

  describe('shouldRetry', () => {
    it('returns true for retryable errors within max attempts', () => {
      const retryableError: AppError = {
        id: 'test',
        type: ErrorType.NETWORK,
        message: 'Network error',
        userMessage: '네트워크 오류',
        timestamp: new Date(),
        retry: {
          canRetry: true,
          attempts: 1,
          maxAttempts: 3,
          nextRetryAt: new Date()
        }
      }
      
      expect(ErrorHandler.shouldRetry(retryableError)).toBe(true)
    })

    it('returns false for non-retryable errors', () => {
      const nonRetryableError: AppError = {
        id: 'test',
        type: ErrorType.VALIDATION,
        message: 'Validation error',
        userMessage: '입력값 오류',
        timestamp: new Date(),
        retry: {
          canRetry: false,
          attempts: 0,
          maxAttempts: 0,
          nextRetryAt: new Date()
        }
      }
      
      expect(ErrorHandler.shouldRetry(nonRetryableError)).toBe(false)
    })

    it('returns false when max attempts reached', () => {
      const maxAttemptsError: AppError = {
        id: 'test',
        type: ErrorType.NETWORK,
        message: 'Network error',
        userMessage: '네트워크 오류',
        timestamp: new Date(),
        retry: {
          canRetry: true,
          attempts: 3,
          maxAttempts: 3,
          nextRetryAt: new Date()
        }
      }
      
      expect(ErrorHandler.shouldRetry(maxAttemptsError)).toBe(false)
    })
  })

  describe('incrementRetryAttempt', () => {
    it('increments attempt count and updates next retry time', () => {
      const error: AppError = {
        id: 'test',
        type: ErrorType.NETWORK,
        message: 'Network error',
        userMessage: '네트워크 오류',
        timestamp: new Date(),
        retry: {
          canRetry: true,
          attempts: 1,
          maxAttempts: 3,
          nextRetryAt: new Date()
        }
      }
      
      const originalNextRetryAt = error.retry.nextRetryAt
      const updatedError = ErrorHandler.incrementRetryAttempt(error)
      
      expect(updatedError.retry.attempts).toBe(2)
      expect(updatedError.retry.nextRetryAt.getTime()).toBeGreaterThan(originalNextRetryAt.getTime())
    })

    it('maintains max attempts limit', () => {
      const error: AppError = {
        id: 'test',
        type: ErrorType.NETWORK,
        message: 'Network error',
        userMessage: '네트워크 오류',
        timestamp: new Date(),
        retry: {
          canRetry: true,
          attempts: 3,
          maxAttempts: 3,
          nextRetryAt: new Date()
        }
      }
      
      const updatedError = ErrorHandler.incrementRetryAttempt(error)
      
      expect(updatedError.retry.attempts).toBe(3) // Should not exceed max
    })
  })

  describe('isRecoverable', () => {
    it('returns true for network errors', () => {
      const networkError: AppError = {
        id: 'test',
        type: ErrorType.NETWORK,
        message: 'Network error',
        userMessage: '네트워크 오류',
        timestamp: new Date(),
        retry: { canRetry: true, attempts: 0, maxAttempts: 3, nextRetryAt: new Date() }
      }
      
      expect(ErrorHandler.isRecoverable(networkError)).toBe(true)
    })

    it('returns true for server errors', () => {
      const serverError: AppError = {
        id: 'test',
        type: ErrorType.SERVER,
        message: 'Server error',
        userMessage: '서버 오류',
        timestamp: new Date(),
        retry: { canRetry: true, attempts: 0, maxAttempts: 3, nextRetryAt: new Date() }
      }
      
      expect(ErrorHandler.isRecoverable(serverError)).toBe(true)
    })

    it('returns false for validation errors', () => {
      const validationError: AppError = {
        id: 'test',
        type: ErrorType.VALIDATION,
        message: 'Validation error',
        userMessage: '입력값 오류',
        timestamp: new Date(),
        retry: { canRetry: false, attempts: 0, maxAttempts: 0, nextRetryAt: new Date() }
      }
      
      expect(ErrorHandler.isRecoverable(validationError)).toBe(false)
    })
  })

  describe('logError', () => {
    it('logs errors with proper format', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
      
      const error: AppError = {
        id: 'test',
        type: ErrorType.NETWORK,
        message: 'Network error',
        userMessage: '네트워크 오류',
        timestamp: new Date(),
        context: 'Test context',
        retry: { canRetry: true, attempts: 0, maxAttempts: 3, nextRetryAt: new Date() }
      }
      
      ErrorHandler.logError(error, 'Test component')
      
      expect(consoleSpy).toHaveBeenCalledWith(
        '[ErrorHandler]',
        'Test component:',
        error
      )
      
      consoleSpy.mockRestore()
    })
  })
})