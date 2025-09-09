import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { Header } from '../../components/Header'
import { AuthContext } from '../../contexts/AuthContext'
import { AuthContextType, AuthUser } from '../../types/auth'

// Mock Next.js router
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(() => ({ push: mockPush })),
}))

// Mock Supabase client
const mockSignOut = jest.fn()
jest.mock('../../lib/supabase/client', () => ({
  supabase: {
    auth: {
      signOut: mockSignOut,
    }
  }
}))

// Mock localStorage and sessionStorage
const mockLocalStorageRemove = jest.fn()
const mockSessionStorageRemove = jest.fn()
Object.defineProperty(window, 'localStorage', {
  value: {
    removeItem: mockLocalStorageRemove,
    getItem: jest.fn(),
    setItem: jest.fn(),
  }
})
Object.defineProperty(window, 'sessionStorage', {
  value: {
    removeItem: mockSessionStorageRemove,
  }
})

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

describe('Logout Flow Integration Test', () => {
  let queryClient: QueryClient
  let mockAuthContext: AuthContextType

  beforeEach(() => {
    jest.clearAllMocks()
    
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })
    
    mockSignOut.mockResolvedValue({ error: null })
    
    mockAuthContext = {
      user: mockUser,
      session: { user: mockUser, access_token: 'token', refresh_token: 'refresh', expires_at: 1234567890 } as any,
      loading: false,
      error: null,
      signInWithGitHub: jest.fn(),
      signOut: jest.fn().mockResolvedValue({}),
      signOutWithConfirmation: jest.fn(),
      refreshSession: jest.fn(),
      isAuthenticated: true,
    }
  })

  const renderWithProviders = (ui: React.ReactElement) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={mockAuthContext}>
          {ui}
        </AuthContext.Provider>
      </QueryClientProvider>
    )
  }

  it('completes full logout flow when user confirms', async () => {
    renderWithProviders(<Header title="Test Dashboard" />)
    
    // Step 1: Find and click logout button
    const logoutButton = screen.getByRole('button', { name: /logout/i })
    expect(logoutButton).toBeInTheDocument()
    
    fireEvent.click(logoutButton)
    
    // Step 2: Confirmation dialog appears
    await waitFor(() => {
      expect(screen.getByText('Confirm Logout')).toBeInTheDocument()
    })
    
    expect(screen.getByText(/are you sure you want to sign out/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign out/i })).toBeInTheDocument()
    
    // Step 3: Confirm logout
    const confirmButton = screen.getByRole('button', { name: /sign out/i })
    fireEvent.click(confirmButton)
    
    // Step 4: Verify signOut was called
    await waitFor(() => {
      expect(mockAuthContext.signOut).toHaveBeenCalledTimes(1)
    })
  })

  it('cancels logout flow when user cancels', async () => {
    renderWithProviders(<Header title="Test Dashboard" />)
    
    // Open logout dialog
    const logoutButton = screen.getByRole('button', { name: /logout/i })
    fireEvent.click(logoutButton)
    
    await waitFor(() => {
      expect(screen.getByText('Confirm Logout')).toBeInTheDocument()
    })
    
    // Cancel logout
    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    fireEvent.click(cancelButton)
    
    // Verify dialog closes and signOut is not called
    await waitFor(() => {
      expect(screen.queryByText('Confirm Logout')).not.toBeInTheDocument()
    })
    
    expect(mockAuthContext.signOut).not.toHaveBeenCalled()
  })

  it('shows user email and logout button in header', () => {
    renderWithProviders(<Header title="Test Dashboard" />)
    
    // User info should be visible
    expect(screen.getByText('test@example.com')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
    
    // Sign in button should not be visible
    expect(screen.queryByRole('button', { name: /sign in/i })).not.toBeInTheDocument()
  })

  it('shows sign in button when not authenticated', () => {
    const unauthenticatedContext = {
      ...mockAuthContext,
      user: null,
      session: null,
      isAuthenticated: false,
    }
    
    render(
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={unauthenticatedContext}>
          <Header title="Test Dashboard" />
        </AuthContext.Provider>
      </QueryClientProvider>
    )
    
    // Sign in button should be visible
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    
    // User info and logout should not be visible
    expect(screen.queryByText('test@example.com')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /logout/i })).not.toBeInTheDocument()
  })

  it('header works with different props combinations', () => {
    const mockRefresh = jest.fn()
    
    renderWithProviders(
      <Header 
        title="Custom Dashboard" 
        onRefresh={mockRefresh}
        isLoading={false}
        showSettings={true}
      >
        <button>Custom Action</button>
      </Header>
    )
    
    // All elements should be present
    expect(screen.getByText('Custom Dashboard')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /refresh/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /settings/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /custom action/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument()
    
    // Test refresh functionality
    const refreshButton = screen.getByRole('button', { name: /refresh/i })
    fireEvent.click(refreshButton)
    expect(mockRefresh).toHaveBeenCalledTimes(1)
  })
})