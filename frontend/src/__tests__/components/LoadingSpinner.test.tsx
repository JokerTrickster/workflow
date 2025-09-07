import React from 'react'
import { render, screen } from '@testing-library/react'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { vi, describe, it, expect } from 'vitest'

// Mock Lucide React icons
vi.mock('lucide-react', () => ({
  Loader2: ({ className }: { className?: string }) => (
    <div data-testid="loader2-icon" className={className}>Spinning...</div>
  ),
}))

describe('LoadingSpinner', () => {
  it('renders with default props', () => {
    render(<LoadingSpinner />)
    
    const icon = screen.getByTestId('loader2-icon')
    expect(icon).toBeInTheDocument()
    expect(icon).toHaveClass('w-6', 'h-6') // Medium size by default
  })

  it('renders with custom message', () => {
    render(<LoadingSpinner message="Loading data..." />)
    
    expect(screen.getByText('Loading data...')).toBeInTheDocument()
  })

  it('renders without message when not provided', () => {
    const { container } = render(<LoadingSpinner />)
    
    const messageElement = container.querySelector('p')
    expect(messageElement).not.toBeInTheDocument()
  })

  it('applies correct size classes', () => {
    const { rerender } = render(<LoadingSpinner size="sm" />)
    let icon = screen.getByTestId('loader2-icon')
    expect(icon).toHaveClass('w-4', 'h-4')
    
    rerender(<LoadingSpinner size="md" />)
    icon = screen.getByTestId('loader2-icon')
    expect(icon).toHaveClass('w-6', 'h-6')
    
    rerender(<LoadingSpinner size="lg" />)
    icon = screen.getByTestId('loader2-icon')
    expect(icon).toHaveClass('w-8', 'h-8')
    
    rerender(<LoadingSpinner size="xl" />)
    icon = screen.getByTestId('loader2-icon')
    expect(icon).toHaveClass('w-12', 'h-12')
  })

  it('applies correct message text sizes based on spinner size', () => {
    const { rerender } = render(<LoadingSpinner size="sm" message="Loading..." />)
    let message = screen.getByText('Loading...')
    expect(message).toHaveClass('text-xs')
    
    rerender(<LoadingSpinner size="md" message="Loading..." />)
    message = screen.getByText('Loading...')
    expect(message).toHaveClass('text-sm')
    
    rerender(<LoadingSpinner size="lg" message="Loading..." />)
    message = screen.getByText('Loading...')
    expect(message).toHaveClass('text-base')
    
    rerender(<LoadingSpinner size="xl" message="Loading..." />)
    message = screen.getByText('Loading...')
    expect(message).toHaveClass('text-lg')
  })

  it('applies custom className', () => {
    const { container } = render(<LoadingSpinner className="custom-spinner" />)
    
    const spinnerContainer = container.firstChild as HTMLElement
    expect(spinnerContainer).toHaveClass('custom-spinner')
  })

  it('renders in fullscreen mode', () => {
    render(<LoadingSpinner fullscreen={true} />)
    
    const fullscreenContainer = screen.getByTestId('loader2-icon').closest('.fixed')
    expect(fullscreenContainer).toBeInTheDocument()
    expect(fullscreenContainer).toHaveClass('fixed', 'inset-0', 'bg-background/80', 'backdrop-blur-sm')
  })

  it('does not render fullscreen overlay when fullscreen is false', () => {
    const { container } = render(<LoadingSpinner fullscreen={false} />)
    
    const fullscreenContainer = container.querySelector('.fixed')
    expect(fullscreenContainer).not.toBeInTheDocument()
  })

  it('combines fullscreen with custom message', () => {
    render(<LoadingSpinner fullscreen={true} message="Processing..." />)
    
    expect(screen.getByText('Processing...')).toBeInTheDocument()
    const fullscreenContainer = screen.getByTestId('loader2-icon').closest('.fixed')
    expect(fullscreenContainer).toBeInTheDocument()
  })

  it('has proper structure and styling', () => {
    const { container } = render(<LoadingSpinner message="Loading..." />)
    
    const spinnerContainer = container.firstChild as HTMLElement
    expect(spinnerContainer).toHaveClass('flex', 'flex-col', 'items-center', 'justify-center', 'gap-3')
    
    const icon = screen.getByTestId('loader2-icon')
    expect(icon).toHaveClass('animate-spin', 'text-primary')
    
    const message = screen.getByText('Loading...')
    expect(message).toHaveClass('text-muted-foreground', 'font-medium')
  })

  it('applies min-h-screen when fullscreen is enabled', () => {
    const { container } = render(<LoadingSpinner fullscreen={true} />)
    
    const innerContainer = container.querySelector('.min-h-screen')
    expect(innerContainer).toBeInTheDocument()
  })

  it('does not apply min-h-screen when fullscreen is disabled', () => {
    const { container } = render(<LoadingSpinner />)
    
    const innerContainer = container.querySelector('.min-h-screen')
    expect(innerContainer).not.toBeInTheDocument()
  })

  it('renders with z-index for fullscreen overlay', () => {
    render(<LoadingSpinner fullscreen={true} />)
    
    const fullscreenContainer = screen.getByTestId('loader2-icon').closest('.fixed')
    expect(fullscreenContainer).toHaveClass('z-50')
  })
})