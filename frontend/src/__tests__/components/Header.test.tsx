import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { Header } from '../../components/Header'
import { AuthContext } from '../../contexts/AuthContext'
import { AuthContextType, AuthUser } from '../../types/auth'

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))

// Mock LogoutButton component
jest.mock('../../components/LogoutButton', () => ({
  LogoutButton: () => <button>Mock Logout</button>
}))

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

describe('Header', () => {
  const mockPush = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    })
  })

  it('renders header with title', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<Header title="Test Dashboard" />, mockAuthContext)
    
    expect(screen.getByText('Test Dashboard')).toBeInTheDocument()
    expect(screen.getByRole('img')).toBeInTheDocument() // GitHub icon
  })

  it('shows user email and logout button when authenticated', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<Header title="Dashboard" />, mockAuthContext)
    
    expect(screen.getByText('test@example.com')).toBeInTheDocument()
    expect(screen.getByText('Mock Logout')).toBeInTheDocument()
  })

  it('shows sign in button when not authenticated', () => {
    const mockAuthContext = createMockAuthContext({
      user: null,
      isAuthenticated: false,
    })
    
    renderWithProviders(<Header title="Dashboard" />, mockAuthContext)
    
    expect(screen.queryByText('test@example.com')).not.toBeInTheDocument()
    expect(screen.queryByText('Mock Logout')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('renders refresh button when onRefresh is provided', () => {
    const mockAuthContext = createMockAuthContext()
    const mockOnRefresh = jest.fn()
    
    renderWithProviders(
      <Header title="Dashboard" onRefresh={mockOnRefresh} />, 
      mockAuthContext
    )
    
    const refreshButton = screen.getByRole('button', { name: /refresh/i })
    expect(refreshButton).toBeInTheDocument()
    
    fireEvent.click(refreshButton)
    expect(mockOnRefresh).toHaveBeenCalledTimes(1)
  })

  it('shows loading state on refresh button', () => {
    const mockAuthContext = createMockAuthContext()
    const mockOnRefresh = jest.fn()
    
    renderWithProviders(
      <Header title="Dashboard" onRefresh={mockOnRefresh} isLoading={true} />, 
      mockAuthContext
    )
    
    const refreshButton = screen.getByRole('button', { name: /refresh/i })
    expect(refreshButton).toBeDisabled()
    
    // Check for loading animation class
    const refreshIcon = refreshButton.querySelector('svg')
    expect(refreshIcon).toHaveClass('animate-spin')
  })

  it('renders settings button when showSettings is true', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(
      <Header title="Dashboard" showSettings={true} />, 
      mockAuthContext
    )
    
    expect(screen.getByRole('button', { name: /settings/i })).toBeInTheDocument()
  })

  it('does not render settings button when showSettings is false', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(
      <Header title="Dashboard" showSettings={false} />, 
      mockAuthContext
    )
    
    expect(screen.queryByRole('button', { name: /settings/i })).not.toBeInTheDocument()
  })

  it('renders custom children', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(
      <Header title="Dashboard">
        <button>Custom Action</button>
      </Header>, 
      mockAuthContext
    )
    
    expect(screen.getByRole('button', { name: /custom action/i })).toBeInTheDocument()
  })

  it('redirects to login when sign in button is clicked', () => {
    const mockAuthContext = createMockAuthContext({
      user: null,
      isAuthenticated: false,
    })
    
    // Mock window.location.href
    delete (window as any).location
    window.location = { href: '' } as any
    
    renderWithProviders(<Header title="Dashboard" />, mockAuthContext)
    
    const signInButton = screen.getByRole('button', { name: /sign in/i })
    fireEvent.click(signInButton)
    
    expect(window.location.href).toBe('/login')
  })

  it('hides user email on small screens', () => {
    const mockAuthContext = createMockAuthContext()
    
    renderWithProviders(<Header title="Dashboard" />, mockAuthContext)
    
    const userEmailElement = screen.getByText('test@example.com')
    expect(userEmailElement).toHaveClass('hidden', 'sm:block')
  })
})