/**
 * @jest-environment jsdom
 */

import React from 'react'
import { render, screen, act, waitFor } from '@testing-library/react'
import { AuthProvider, useAuth } from '@/contexts/AuthContext'

// Mock Supabase client
jest.mock('@/lib/supabase/client', () => {
  const mockSupabase = {
    auth: {
      getSession: jest.fn(),
      onAuthStateChange: jest.fn(),
      signInWithOAuth: jest.fn(),
      signOut: jest.fn(),
      refreshSession: jest.fn()
    }
  }
  return { supabase: mockSupabase }
})

import { supabase } from '@/lib/supabase/client'
const mockSupabase = supabase as jest.Mocked<typeof supabase>

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    origin: 'http://localhost:3000'
  },
  writable: true
})

// Test component that uses the auth context
function TestComponent() {
  const {
    user,
    loading,
    error,
    signInWithGitHub,
    signOut,
    refreshSession,
    isAuthenticated
  } = useAuth()

  return (
    <div>
      <div data-testid="loading">{loading.toString()}</div>
      <div data-testid="authenticated">{isAuthenticated.toString()}</div>
      <div data-testid="user">{user ? user.email : 'null'}</div>
      <div data-testid="error">{error || 'null'}</div>
      <button onClick={signInWithGitHub} data-testid="signin">
        Sign In
      </button>
      <button onClick={signOut} data-testid="signout">
        Sign Out
      </button>
      <button onClick={refreshSession} data-testid="refresh">
        Refresh
      </button>
    </div>
  )
}

describe('AuthContext', () => {
  let mockUnsubscribe: jest.Mock

  beforeEach(() => {
    jest.clearAllMocks()
    mockUnsubscribe = jest.fn()
    
    // Default mock for onAuthStateChange
    mockSupabase.auth.onAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: mockUnsubscribe } }
    })
    
    // Mock console.error and console.log to keep test output clean
    jest.spyOn(console, 'error').mockImplementation(() => {})
    jest.spyOn(console, 'log').mockImplementation(() => {})
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  it('should provide auth context to child components', async () => {
    // Mock initial session call
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    // Initially loading should be true
    expect(screen.getByTestId('loading')).toHaveTextContent('true')

    // Wait for loading to complete
    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('authenticated')).toHaveTextContent('false')
    expect(screen.getByTestId('user')).toHaveTextContent('null')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
  })

  it('should initialize with existing session', async () => {
    const mockSession = {
      access_token: 'token',
      user: { id: 'user-id', email: 'test@example.com' }
    }

    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: mockSession },
      error: null
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('authenticated')).toHaveTextContent('true')
    expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
  })

  it('should handle initialization error', async () => {
    const mockError = { message: 'Session retrieval failed' }
    
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: mockError
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('error')).toHaveTextContent('Session retrieval failed')
  })

  it('should handle auth state changes', async () => {
    // Mock initial session call
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null
    })

    let authStateCallback: (event: string, session: unknown) => void

    // Capture the callback passed to onAuthStateChange
    mockSupabase.auth.onAuthStateChange.mockImplementation((callback) => {
      authStateCallback = callback
      return { data: { subscription: { unsubscribe: mockUnsubscribe } } }
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    // Simulate sign in event
    const mockSession = {
      access_token: 'token',
      user: { id: 'user-id', email: 'test@example.com' }
    }

    await act(async () => {
      authStateCallback!('SIGNED_IN', mockSession)
    })

    expect(screen.getByTestId('authenticated')).toHaveTextContent('true')
    expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
  })

  it('should handle GitHub sign in', async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null
    })

    mockSupabase.auth.signInWithOAuth.mockResolvedValue({
      error: null
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await act(async () => {
      screen.getByTestId('signin').click()
    })

    expect(mockSupabase.auth.signInWithOAuth).toHaveBeenCalledWith({
      provider: 'github',
      options: {
        redirectTo: 'http://localhost:3000/auth/callback'
      }
    })
  })

  it('should handle sign in error', async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null
    })

    const mockError = { message: 'OAuth failed' }
    mockSupabase.auth.signInWithOAuth.mockResolvedValue({
      error: mockError
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await act(async () => {
      screen.getByTestId('signin').click()
    })

    expect(screen.getByTestId('error')).toHaveTextContent('OAuth failed')
  })

  it('should handle sign out', async () => {
    const mockSession = {
      access_token: 'token',
      user: { id: 'user-id', email: 'test@example.com' }
    }

    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: mockSession },
      error: null
    })

    mockSupabase.auth.signOut.mockResolvedValue({
      error: null
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('authenticated')).toHaveTextContent('true')
    })

    await act(async () => {
      screen.getByTestId('signout').click()
    })

    expect(mockSupabase.auth.signOut).toHaveBeenCalled()
  })

  it('should handle session refresh', async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null
    })

    const mockRefreshedSession = {
      access_token: 'new-token',
      user: { id: 'user-id', email: 'test@example.com' }
    }

    mockSupabase.auth.refreshSession.mockResolvedValue({
      data: { session: mockRefreshedSession },
      error: null
    })

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await act(async () => {
      screen.getByTestId('refresh').click()
    })

    expect(mockSupabase.auth.refreshSession).toHaveBeenCalled()
  })

  it('should throw error when useAuth is used outside provider', () => {
    // Mock console.error to prevent error output in test
    const mockConsoleError = jest.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      render(<TestComponent />)
    }).toThrow('useAuth must be used within an AuthProvider')

    mockConsoleError.mockRestore()
  })

  it('should unsubscribe from auth changes on unmount', async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null
    })

    const { unmount } = render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    unmount()

    expect(mockUnsubscribe).toHaveBeenCalled()
  })
})