import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { GitHubButton } from '@/components/GitHubButton'
import { vi, describe, it, expect, beforeEach } from 'vitest'

// Mock Lucide React icons
vi.mock('lucide-react', () => ({
  Loader2: ({ className }: { className?: string }) => (
    <div data-testid="loader-icon" className={className}>Loading...</div>
  ),
}))

describe('GitHubButton', () => {
  const mockOnClick = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with default props', () => {
    render(<GitHubButton onClick={mockOnClick} />)
    
    const button = screen.getByRole('button', { name: /sign in with github/i })
    expect(button).toBeInTheDocument()
    expect(screen.getByText('Continue with GitHub')).toBeInTheDocument()
    
    // Check for GitHub icon (SVG)
    const githubIcon = button.querySelector('svg')
    expect(githubIcon).toBeInTheDocument()
  })

  it('renders custom children text', () => {
    render(<GitHubButton onClick={mockOnClick}>Sign in with GitHub</GitHubButton>)
    
    expect(screen.getByText('Sign in with GitHub')).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    render(<GitHubButton onClick={mockOnClick} />)
    
    const button = screen.getByRole('button')
    fireEvent.click(button)
    
    expect(mockOnClick).toHaveBeenCalledTimes(1)
  })

  it('shows loading state when loading prop is true', () => {
    render(<GitHubButton onClick={mockOnClick} loading={true} />)
    
    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
    expect(screen.getByTestId('loader-icon')).toBeInTheDocument()
    expect(screen.getByText('Signing in...')).toBeInTheDocument()
  })

  it('shows loading state during onClick execution', async () => {
    const slowOnClick = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)))
    render(<GitHubButton onClick={slowOnClick} />)
    
    const button = screen.getByRole('button')
    fireEvent.click(button)
    
    // Button should immediately show loading state
    expect(screen.getByTestId('loader-icon')).toBeInTheDocument()
    expect(screen.getByText('Signing in...')).toBeInTheDocument()
    expect(button).toBeDisabled()
    
    // Wait for onClick to resolve
    await waitFor(() => {
      expect(slowOnClick).toHaveBeenCalledTimes(1)
    })
  })

  it('is disabled when disabled prop is true', () => {
    render(<GitHubButton onClick={mockOnClick} disabled={true} />)
    
    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
    
    fireEvent.click(button)
    expect(mockOnClick).not.toHaveBeenCalled()
  })

  it('does not call onClick when disabled', () => {
    render(<GitHubButton onClick={mockOnClick} disabled={true} />)
    
    const button = screen.getByRole('button')
    fireEvent.click(button)
    
    expect(mockOnClick).not.toHaveBeenCalled()
  })

  it('does not call onClick when loading', () => {
    render(<GitHubButton onClick={mockOnClick} loading={true} />)
    
    const button = screen.getByRole('button')
    fireEvent.click(button)
    
    expect(mockOnClick).not.toHaveBeenCalled()
  })

  it('applies custom className', () => {
    render(<GitHubButton onClick={mockOnClick} className="custom-class" />)
    
    const button = screen.getByRole('button')
    expect(button).toHaveClass('custom-class')
  })

  it('applies different sizes', () => {
    const { rerender } = render(<GitHubButton onClick={mockOnClick} size="sm" />)
    let button = screen.getByRole('button')
    expect(button).toHaveClass('h-8') // Small size class
    
    rerender(<GitHubButton onClick={mockOnClick} size="lg" />)
    button = screen.getByRole('button')
    expect(button).toHaveClass('h-10') // Large size class
  })

  it('applies different variants', () => {
    const { rerender } = render(<GitHubButton onClick={mockOnClick} variant="default" />)
    let button = screen.getByRole('button')
    expect(button).toHaveClass('bg-gray-900') // Default variant styling
    
    rerender(<GitHubButton onClick={mockOnClick} variant="outline" />)
    button = screen.getByRole('button')
    expect(button).toHaveClass('border-gray-300') // Outline variant styling
  })

  it('handles async onClick errors gracefully', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    const errorOnClick = vi.fn(() => Promise.reject(new Error('Test error')))
    
    render(<GitHubButton onClick={errorOnClick} />)
    
    const button = screen.getByRole('button')
    fireEvent.click(button)
    
    await waitFor(() => {
      expect(errorOnClick).toHaveBeenCalledTimes(1)
      expect(consoleError).toHaveBeenCalledWith('GitHub sign-in error:', expect.any(Error))
    })
    
    consoleError.mockRestore()
  })

  it('has proper accessibility attributes', () => {
    render(<GitHubButton onClick={mockOnClick} />)
    
    const button = screen.getByRole('button')
    expect(button).toHaveAttribute('aria-label', 'Sign in with GitHub')
  })

  it('prevents double clicks during loading', async () => {
    const slowOnClick = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)))
    render(<GitHubButton onClick={slowOnClick} />)
    
    const button = screen.getByRole('button')
    
    // Click multiple times rapidly
    fireEvent.click(button)
    fireEvent.click(button)
    fireEvent.click(button)
    
    // Should only be called once
    await waitFor(() => {
      expect(slowOnClick).toHaveBeenCalledTimes(1)
    })
  })

  it('renders GitHub icon SVG with correct viewBox', () => {
    render(<GitHubButton onClick={mockOnClick} />)
    
    const button = screen.getByRole('button')
    const svgIcon = button.querySelector('svg')
    
    expect(svgIcon).toBeInTheDocument()
    expect(svgIcon).toHaveAttribute('viewBox', '0 0 16 16')
    expect(svgIcon).toHaveAttribute('fill', 'currentColor')
  })
})