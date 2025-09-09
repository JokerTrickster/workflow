import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { LogoutButton } from '../../components/LogoutButton'
import { AuthContext } from '../../contexts/AuthContext'
import { AuthContextType, AuthUser } from '../../types/auth'

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))

// Mock user data
const mockUser: AuthUser = {
  id: 'test-user-id',
  email: 'test@example.com',
  aud: 'authenticated',
  role: 'authenticated',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  last_sign_in_at: '2024-01-01T00:00:00Z',
  app_metadata: {},
  user_metadata: {},
}

const createMockAuthContext = (overrides: Partial<AuthContextType> = {}): AuthContextType => ({
  user: mockUser,
  session: { user: mockUser, access_token: 'token', refresh_token: 'refresh', expires_at: 1234567890 } as any,
  loading: false,
  error: null,
  signInWithGitHub: jest.fn(),
  signOut: jest.fn().mockResolvedValue({}),
  signOutWithConfirmation: jest.fn(),
  refreshSession: jest.fn(),
  isAuthenticated: true,
  ...overrides,
})

const renderWithProviders = (ui: React.ReactElement, authContext: AuthContextType) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={authContext}>
        {ui}
      </AuthContext.Provider>
    </QueryClientProvider>
  )
}

describe('LogoutButton', () => {
  const mockPush = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    })
  })

  it('renders logout button when user is authenticated', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<LogoutButton />, mockAuthContext)
    
    expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
  })

  it('does not render when user is not authenticated', () => {
    const mockAuthContext = createMockAuthContext({
      user: null,
      isAuthenticated: false,
    })
    
    renderWithProviders(<LogoutButton />, mockAuthContext)
    
    expect(screen.queryByRole('button', { name: /logout/i })).not.toBeInTheDocument()
  })

  it('shows confirmation dialog when logout button is clicked', async () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<LogoutButton />, mockAuthContext)
    
    const logoutButton = screen.getByRole('button', { name: /logout/i })
    fireEvent.click(logoutButton)
    
    await waitFor(() => {
      expect(screen.getByText('Confirm Logout')).toBeInTheDocument()
    })
    
    expect(screen.getByText(/are you sure you want to sign out/i)).toBeInTheDocument()
  })

  it('cancels logout when cancel button is clicked', async () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<LogoutButton />, mockAuthContext)
    
    // Open dialog
    const logoutButton = screen.getByRole('button', { name: /logout/i })
    fireEvent.click(logoutButton)
    
    await waitFor(() => {
      expect(screen.getByText('Confirm Logout')).toBeInTheDocument()
    })
    
    // Cancel logout
    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    fireEvent.click(cancelButton)
    
    await waitFor(() => {
      expect(screen.queryByText('Confirm Logout')).not.toBeInTheDocument()
    })
    
    expect(mockAuthContext.signOut).not.toHaveBeenCalled()
  })

  it('calls signOut when confirmed', async () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<LogoutButton />, mockAuthContext)
    
    // Open dialog
    const logoutButton = screen.getByRole('button', { name: /logout/i })
    fireEvent.click(logoutButton)
    
    await waitFor(() => {
      expect(screen.getByText('Confirm Logout')).toBeInTheDocument()
    })
    
    // Confirm logout
    const signOutButton = screen.getByRole('button', { name: /sign out/i })
    fireEvent.click(signOutButton)
    
    await waitFor(() => {
      expect(mockAuthContext.signOut).toHaveBeenCalledTimes(1)
    })
  })

  it('shows loading state during logout', async () => {
    const mockSignOut = jest.fn().mockImplementation(
      () => new Promise(resolve => setTimeout(resolve, 100))
    )
    const mockAuthContext = createMockAuthContext({
      signOut: mockSignOut,
    })
    
    renderWithProviders(<LogoutButton />, mockAuthContext)
    
    // Open dialog and confirm
    const logoutButton = screen.getByRole('button', { name: /logout/i })
    fireEvent.click(logoutButton)
    
    await waitFor(() => {
      expect(screen.getByText('Confirm Logout')).toBeInTheDocument()
    })
    
    const signOutButton = screen.getByRole('button', { name: /sign out/i })
    fireEvent.click(signOutButton)
    
    // Check loading state
    expect(signOutButton).toBeDisabled()
    expect(screen.getByText('Sign Out')).toBeInTheDocument()
    
    await waitFor(() => {
      expect(mockSignOut).toHaveBeenCalledTimes(1)
    })
  })

  it('can render without text', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<LogoutButton showText={false} />, mockAuthContext)
    
    const logoutButton = screen.getByRole('button')
    expect(logoutButton).toBeInTheDocument()
    expect(logoutButton).not.toHaveTextContent('Logout')
  })

  it('applies custom className and variant', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(
      <LogoutButton className="custom-class" variant="destructive" />, 
      mockAuthContext
    )
    
    const logoutButton = screen.getByRole('button')
    expect(logoutButton).toHaveClass('custom-class')
  })
})