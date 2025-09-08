import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorMessage } from '../../components/ErrorMessage'
import { AppError, ErrorType } from '../../utils/errorHandler'

const mockError: AppError = {
  id: 'test-error',
  type: ErrorType.NETWORK,
  message: 'Network connection failed',
  userMessage: '네트워크 연결에 실패했습니다',
  timestamp: new Date(),
  context: 'Test context',
  retry: {
    canRetry: true,
    attempts: 1,
    maxAttempts: 3,
    nextRetryAt: new Date(Date.now() + 5000)
  }
}

const mockOnRetry = jest.fn()

describe('ErrorMessage Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders error message with user-friendly text', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} />)
    
    expect(screen.getByText('네트워크 연결에 실패했습니다')).toBeInTheDocument()
  })

  it('shows retry button when error is retryable', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} />)
    
    const retryButton = screen.getByRole('button', { name: /다시 시도/i })
    expect(retryButton).toBeInTheDocument()
  })

  it('calls onRetry when retry button is clicked', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} />)
    
    const retryButton = screen.getByRole('button', { name: /다시 시도/i })
    fireEvent.click(retryButton)
    
    expect(mockOnRetry).toHaveBeenCalledTimes(1)
  })

  it('does not show retry button for non-retryable errors', () => {
    const nonRetryableError = {
      ...mockError,
      retry: { ...mockError.retry, canRetry: false }
    }
    
    render(<ErrorMessage error={nonRetryableError} onRetry={mockOnRetry} />)
    
    expect(screen.queryByRole('button', { name: /다시 시도/i })).not.toBeInTheDocument()
  })

  it('shows technical details when showDetails is true', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} showDetails />)
    
    expect(screen.getByText('Network connection failed')).toBeInTheDocument()
    expect(screen.getByText('Test context')).toBeInTheDocument()
  })

  it('does not show technical details when showDetails is false', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} showDetails={false} />)
    
    expect(screen.queryByText('Network connection failed')).not.toBeInTheDocument()
    expect(screen.queryByText('Test context')).not.toBeInTheDocument()
  })

  it('displays retry attempts information', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} showDetails />)
    
    expect(screen.getByText(/재시도: 1\/3/)).toBeInTheDocument()
  })

  it('renders with appropriate ARIA attributes for accessibility', () => {
    render(<ErrorMessage error={mockError} onRetry={mockOnRetry} />)
    
    const errorContainer = screen.getByRole('alert')
    expect(errorContainer).toBeInTheDocument()
  })
})